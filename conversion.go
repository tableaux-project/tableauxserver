package tableauxserver

import (
	"fmt"

	"github.com/tableaux-project/tableaux/config"
	"github.com/tableaux-project/tableaux/datasource"
)

// MapDataRequestColumns extract the actual TableSchemaColumns from a DataRequest.
func MapDataRequestColumns(dataRequest DataRequest, schema config.ResolvedTableSchema) ([]config.TableSchemaColumn, error) {
	resolvedColumns := make([]config.TableSchemaColumn, len(dataRequest.Columns))

	for i, column := range dataRequest.Columns {
		columnPath := column.Name

		resolvedColumn, err := schema.Column(columnPath)
		if err != nil {
			return nil, fmt.Errorf("unknown column %s", columnPath)
		}

		resolvedColumns[i] = resolvedColumn
	}

	return resolvedColumns, nil
}

// MapDataRequestOrders maps the ColumnOrders of a DataRequest to abstract tableaux Orders.
func MapDataRequestOrders(dataRequest DataRequest, schema config.ResolvedTableSchema) ([]datasource.Order, error) {
	orders := make([]datasource.Order, len(dataRequest.Order))

	for i, order := range dataRequest.Order {
		columnPath := order.Column

		if _, err := schema.Column(columnPath); err == config.ErrUnknownColumn {
			return nil, fmt.Errorf("unknown order column %s", columnPath)
		}

		orders[i] = datasource.NewOrder(order.Column, order.Dir, nil)
	}

	return orders, nil
}

// MapDataRequestFilters maps the Search values of the Columns of a DataRequest to abstract tableaux FilterGroups.
func MapDataRequestFilters(dataRequest DataRequest) ([]datasource.FilterGroup, error) {
	var filterGroups []datasource.FilterGroup

	for _, column := range dataRequest.Columns {
		if len(column.Search) > 0 {
			filters := make([]datasource.Filter, len(column.Search))

			for i, search := range column.Search {
				filters[i] = datasource.NewFilter(search.Mode, search.Value)
			}

			filterGroups = append(filterGroups, datasource.NewFilterGroup(column.Name, filters))
		}
	}

	return filterGroups, nil
}
