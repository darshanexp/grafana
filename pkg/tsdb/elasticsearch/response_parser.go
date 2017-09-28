package elasticsearch

import (
	"encoding/json"
	"fmt"
	"strconv"
	"strings"

	"github.com/grafana/grafana/pkg/components/null"
	"github.com/grafana/grafana/pkg/tsdb"
)

func joinMaps(left map[string]tsdb.TimeSeriesPoints, right map[string]tsdb.TimeSeriesPoints) map[string]tsdb.TimeSeriesPoints {
	result := map[string]tsdb.TimeSeriesPoints{}
	for key, value := range left {
		result[key] = value
	}
	for key, value := range right {
		if _, ok := result[key]; ok {
			for _, pt := range value {
				result[key] = append(result[key], pt)
			}
		} else {
			result[key] = value
		}
	}

	return result
}

func parseSubQueryResults(parentAggregationKey string, bucketAggregationKey string, bucketlist BucketList, preferredNames NameMap, resultFilter FilterMap) (map[string]tsdb.TimeSeriesPoints, string, error) {
	timeSeries := map[string]tsdb.TimeSeriesPoints{}

	for _, bucket := range bucketlist.Buckets {
		rawAggregation, _ := json.Marshal(bucket)

		aggregations := make(map[string]interface{})
		err := json.Unmarshal(rawAggregation, &aggregations)
		if err != nil {
			return timeSeries, "", err
		}

		metricKey := ""
		docCount := 0.0
		var valueRow [2]null.Float

		if k, ok := aggregations["key"]; ok {
			if v, ok := k.(string); ok {
				bucketAggregationKey = fmt.Sprintf("%s.%s", parentAggregationKey, v)
			}
		}

		for key, value := range aggregations {
			switch value.(type) {
			case string:
				if key == "key_as_string" {
					keyf, err := strconv.ParseFloat(value.(string), 64)
					if err == nil {
						valueRow[1] = parseValue(keyf)
					}
				}
			case float64:
				if key == "key" {
					valueRow[1] = parseValue(value.(float64))
				} else if key == "doc_count" {
					docCount = value.(float64)
				}
			case map[string]interface{}:
				valueMap := value.(map[string]interface{})
				if valueMap["value"] != nil {
					metricKey = key
					valueRow[0] = parseValue(valueMap["value"].(float64))
				} else if valueMap["buckets"] != nil {
					buckets := Bucket{}

					bucketBytes, err := json.Marshal(valueMap["buckets"])
					if err != nil {
						return timeSeries, bucketAggregationKey, err
					}

					err = json.Unmarshal(bucketBytes, &buckets)
					if err != nil {
						return timeSeries, bucketAggregationKey, err
					}

					mykey := key
					if bucketAggregationKey != "" {
						mykey = bucketAggregationKey
						metricKey = mykey
					} else {
						mykey = fmt.Sprintf("%s%s", parentAggregationKey, mykey)
					}
					nestedBucketList := BucketList{
						Buckets: buckets,
					}
					nestedTimeSeries, tBucketKey, err := parseSubQueryResults(mykey, bucketAggregationKey, nestedBucketList, preferredNames, resultFilter)

					if tBucketKey != "" {
						bucketAggregationKey = tBucketKey
					}
					if err != nil {
						return timeSeries, bucketAggregationKey, err
					}

					timeSeries = joinMaps(timeSeries, nestedTimeSeries)
				}
			default:
				fmt.Printf("Unknown Type: %v %v\n", key, value)
			}

			if metricKey != "" {
				name := preferredNames.GetName(metricKey)
				if bucketAggregationKey != "" {
					name = bucketAggregationKey
				}

				if !resultFilter.Hide(metricKey) {
					if _, ok := timeSeries[name]; !ok {
						timeSeries[name] = make(tsdb.TimeSeriesPoints, 0)
					}
					timeSeries[name] = append(timeSeries[name], valueRow)
				}
			}
		}

		if metricKey == "" {
			name := "doc_count"

			if _, ok := timeSeries[name]; !ok {
				timeSeries[name] = make(tsdb.TimeSeriesPoints, 0)
			}
			valueRow[0] = parseValue(docCount)
			timeSeries[name] = append(timeSeries[name], valueRow)
		}
	}

	return timeSeries, "", nil
}

func cleanupName(name string) string {
	parts := strings.Split(name, ".")
	if len(parts) <= 1 {
		return name
	} else {
		return strings.Join(parts[1:], ".")
	}
}

func parseQueryResult(response []byte, preferredNames NameMap, resultFilter FilterMap) (*tsdb.QueryResult, error) {
	queryRes := tsdb.NewQueryResult()

	esSearchResult := &Response{}
	err := json.Unmarshal(response, esSearchResult)
	if err != nil {
		return nil, err
	}

	timeSeries := map[string]tsdb.TimeSeriesPoints{}
	for aggregationID, buckets := range esSearchResult.Aggregations {
		tSeries, _, err := parseSubQueryResults(aggregationID, "", buckets, preferredNames, resultFilter)

		if err != nil {
			return nil, err
		}

		timeSeries = joinMaps(timeSeries, tSeries)
	}

	for id, series := range timeSeries {
		if len(timeSeries) > 0 && id != "doc_count" || len(timeSeries) == 1 && id == "doc_count" {
			// Remove all points that have null data for either coordinate value
			nonNullPoints := make(tsdb.TimeSeriesPoints, 0)
			seenTimes := make(map[float64]bool)
			for _, v := range series {
				if v[0].Ptr() != nil && v[1].Ptr() != nil {
					_, seenTime := seenTimes[v[1].Float64]
					// Discard duplicate timestamps (Elasticsearch seems to return these occasionally). Important to do so before
					// cropping.
					if !seenTime {
						nonNullPoints = append(nonNullPoints, v)
						seenTimes[v[1].Float64] = true
					}
				}
			}
			// Auto-cropping both ends for Riot specific HMP 2.0 per-minute calculations. We only want whole datapoints.
			if len(nonNullPoints) > 1 {
				nonNullPoints = nonNullPoints[1 : len(nonNullPoints)-1]
			}
			ts := &tsdb.TimeSeries{
				Name:   cleanupName(id),
				Points: nonNullPoints,
			}
			queryRes.Series = append(queryRes.Series, ts)
		}
	}

	return queryRes, nil
}

func parseValue(value float64) null.Float {
	return null.FloatFrom(float64(value))
}
