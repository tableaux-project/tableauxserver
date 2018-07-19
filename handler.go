package tableauxserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"

	"github.com/NYTimes/gziphandler"
	"github.com/gorilla/mux"
	"github.com/rs/cors"
	"gopkg.in/birkirb/loggers.v1/log"

	"github.com/tableaux-project/tableaux"
	"github.com/tableaux-project/tableaux/config"
	"github.com/tableaux-project/tableaux/datasource"
)

// NewServer creates a new http Server, with a backing tableaux wrapper for routing data calls.
func NewServer(databaseConnector datasource.Connector, schemaMapper config.SchemaMapper, port int64, enableSimpleGet bool) (http.Server, error) {
	wrapper := tableauxServerWrapper{
		databaseConnector,
		schemaMapper,
	}

	rtr := mux.NewRouter()
	for name := range schemaMapper.ResolvedSchemas() {
		urlPath := fmt.Sprintf("/api/v1/%s/", name)
		log.WithField("endpoint", urlPath).Info("Mapping data endpoint")

		rtr.HandleFunc("/api/v1/{schema}", http.HandlerFunc(wrapper.dataHandler)).Methods("POST")

		if enableSimpleGet {
			rtr.HandleFunc("/api/v1/{schema}", http.HandlerFunc(wrapper.simpleDataHandler)).Methods("GET")
		}
	}

	return http.Server{
		Addr: ":" + strconv.FormatInt(port, 10),
		Handler: gziphandler.GzipHandler(cors.New(cors.Options{
			AllowedMethods: []string{"POST", "GET"},
		}).Handler(rtr)),
	}, nil
}

// ValidateDataRequest does a quick validation if the most important parameters are
// set correctly.
func ValidateDataRequest(dataRequest DataRequest) error {
	if dataRequest.Length <= 0 {
		return errors.New("parameter 'Length' must be positive and greater than 0")
	}
	if dataRequest.Start < 0 {
		return errors.New("parameter 'Start' must be positive and greater or equal to 0")
	}
	if dataRequest.Draw < 0 {
		return errors.New("parameter 'Draw' must be positive and greater or equal to 0")
	}

	if len(dataRequest.Columns) == 0 {
		return errors.New("no columns selected")
	}

	return nil
}

type tableauxServerWrapper struct {
	connector    datasource.Connector
	schemaMapper config.SchemaMapper
}

func (th tableauxServerWrapper) simpleDataHandler(writer http.ResponseWriter, req *http.Request) {
	// We are always responding in json - even in error cases
	writer.Header().Add("Content-Type", "application/json")

	// Resolve the schema to be used
	schema, err := th.schemaMapper.ResolvedSchema(mux.Vars(req)["schema"])
	if err != nil {
		writeErrorResponse(writer, http.StatusNotFound, 0, "Unknown schema")
		return
	}

	// ---------------

	requestColumns := make([]Column, len(schema.Columns()))
	for i, column := range schema.Columns() {
		requestColumns[i] = Column{Name: column.Path}

		if column.Path == "dosemeter_assignedEmployee_awstEmployeeNumber" {
			requestColumns[i].Search = []ColumnSearch{
				{Value: "MA0680872", Mode: tableaux.FilterEquals},
			}
		}
	}

	dataRequest := DataRequest{
		Start:   0,
		Length:  10,
		Columns: requestColumns,
		Locale:  "de",
		Order: []ColumnOrder{
			{
				Column: "dosemeter_assignedEmployee_dosemeters",
				Dir:    tableaux.OrderDesc,
			},
		},
	}

	// ---------------

	th.handleRequest(writer, dataRequest, schema)
}

func (th tableauxServerWrapper) dataHandler(writer http.ResponseWriter, req *http.Request) {
	// We are always responding in json - even in error cases
	writer.Header().Add("Content-Type", "application/json")

	// Resolve the schema to be used
	schema, err := th.schemaMapper.ResolvedSchema(mux.Vars(req)["schema"])
	if err != nil {
		writeErrorResponse(writer, http.StatusNotFound, 0, "Unknown schema")
		return
	}

	// ---------------

	decoder := json.NewDecoder(req.Body)
	var dataRequest DataRequest
	if err := decoder.Decode(&dataRequest); err != nil {
		writeErrorResponse(writer, http.StatusBadRequest, 0, "Failed to parse request")
		return
	}

	// ---------------

	th.handleRequest(writer, dataRequest, schema)
}

func (th tableauxServerWrapper) handleRequest(writer http.ResponseWriter, dataRequest DataRequest, schema config.ResolvedTableSchema) {
	// First - a quick validation
	if err := ValidateDataRequest(dataRequest); err != nil {
		writeErrorResponse(writer, http.StatusBadRequest, dataRequest.Draw, err.Error())
		return
	}

	// ---------------

	// Map from tableaux server to tableaux DTOs

	columns, err := MapDataRequestColumns(dataRequest, schema)
	if err != nil {
		writeErrorResponse(writer, http.StatusBadRequest, dataRequest.Draw, err.Error())
		return
	}

	filters, err := MapDataRequestFilters(dataRequest)
	if err != nil {
		writeErrorResponse(writer, http.StatusBadRequest, dataRequest.Draw, err.Error())
		return
	}

	orders, err := MapDataRequestOrders(dataRequest, schema)
	if err != nil {
		writeErrorResponse(writer, http.StatusBadRequest, dataRequest.Draw, err.Error())
		return
	}

	// Validate if the data source can handle the request
	if valErr := th.connector.ValidateRequest(
		columns,
		schema,
		filters,
		orders,
		dataRequest.Search.Value,
		10,
		0,
		dataRequest.Locale,
	); valErr != nil {
		log.WithField("error", valErr).Error("Failed to validate request")
		writer.WriteHeader(http.StatusBadRequest)
		return
	}

	// Get the data!
	result, totalCount, filteredCount, err := th.connector.FetchData(
		columns,
		schema,
		filters,
		orders,
		dataRequest.Search.Value,
		10,
		0,
		dataRequest.Locale,
	)

	if err != nil {
		writeErrorResponse(writer, http.StatusBadRequest, dataRequest.Draw, err.Error())
		return
	}

	// ---------------

	// At this point, all the data is available. Lets write it to the response.

	// TODO: Error handling must be improved here! This is very fragile!

	writer.Write([]byte(`{"data": [`))

	for i, row := range *result {
		writeRow(columns, row, writer)
		if i < len(*result)-1 {
			writer.Write([]byte(","))
		}
	}

	writer.Write([]byte("],"))

	writer.Write([]byte(fmt.Sprintf(`"draw": %d, `, dataRequest.Draw)))
	writer.Write([]byte(fmt.Sprintf(`"recordsTotal": %d,`, totalCount)))
	writer.Write([]byte(fmt.Sprintf(`"recordsFiltered": %d}`, filteredCount)))
}

func writeErrorResponse(writer http.ResponseWriter, statusCode int, draw int64, message string) {
	writer.WriteHeader(statusCode)
	m, _ := json.Marshal(message)
	writer.Write([]byte(fmt.Sprintf(`{"error": %s, "draw": %d}`, m, draw)))
}

func writeRow(columns []config.TableSchemaColumn, columnData map[string]interface{}, writer io.Writer) {
	writer.Write([]byte("{"))

	count := 0
	for _, column := range columns {
		writer.Write([]byte("\"" + column.Path + "\""))
		writer.Write([]byte(":"))

		key, _ := json.Marshal(columnData[column.Path])
		writer.Write(key)

		count++

		// Not last column
		if count < len(columnData) {
			writer.Write([]byte(","))
		}
	}

	writer.Write([]byte("}"))
}
