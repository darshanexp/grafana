+++
title = "Annotations HTTP API "
description = "Grafana Annotations HTTP API"
keywords = ["grafana", "http", "documentation", "api", "annotation", "annotations", "comment"]
aliases = ["/http_api/annotations/"]
type = "docs"
[menu.docs]
name = "Annotations"
identifier = "annotationshttp"
parent = "http_api"
+++

# Annotations resources / actions

This is the API documentation for the new Grafana Annotations feature released in Grafana 4.6. Annotations are saved in the Grafana database (sqlite, mysql or postgres). Annotations can be global annotations that can be shown on any dashboard by configuring an annotation data source - they are filtered by tags. Or they can be tied to a panel on a dashboard and are then only shown on that panel.

## Find Annotations

`GET /api/annotations?from=1506676478816&to=1507281278816&tags=tag1&tags=tag2&limit=100`

**Example Request**:

```http
GET /api/annotations?from=1506676478816&to=1507281278816&tags=tag1&tags=tag2&limit=100 HTTP/1.1
Accept: application/json
Content-Type: application/json
Authorization: Basic YWRtaW46YWRtaW4=
```


Query Parameters:

- `from`: epoch datetime in milliseconds. Optional.
- `to`: epoch datetime in milliseconds. Optional.
- `limit`: number. Optional - default is 10. Max limit for results returned.
- `alertId`: number. Optional. Find annotations for a specified alert.
- `dashboardId`: number. Optional. Find annotations that are scoped to a specific dashboard
- `panelId`: number. Optional. Find annotations that are scoped to a specific panel
- `tags`: string. Optional. Use this to filter global annotations. Global annotations are annotations from an annotation data source that are not connected specifically to a dashboard or panel. To do an "AND" filtering with multiple tags, specify the tags parameter multiple times e.g. `tags=tag1&tags=tag2`.

**Example Response**:

```http
HTTP/1.1 200
Content-Type: application/json
[
    {
        "id": 1124,
        "alertId": 0,
        "dashboardId": 468,
        "panelId": 2,
        "userId": 1,
        "userName": "",
        "newState": "",
        "prevState": "",
        "time": 1507266395000,
        "text": "test",
        "metric": "",
        "regionId": 1123,
        "type": "event",
        "tags": [
            "tag1",
            "tag2"
        ],
        "data": {}
    },
    {
        "id": 1123,
        "alertId": 0,
        "dashboardId": 468,
        "panelId": 2,
        "userId": 1,
        "userName": "",
        "newState": "",
        "prevState": "",
        "time": 1507265111000,
        "text": "test",
        "metric": "",
        "regionId": 1123,
        "type": "event",
        "tags": [
            "tag1",
            "tag2"
        ],
        "data": {}
    }
]
```

## Create Annotation

Creates an annotation in the Grafana database. The `dashboardId` and `panelId` fields are optional. If they are not specified then a global annotation is created and can be queried in any dashboard that adds the Grafana annotations data source. When creating a region annotation the response will include both `id` and `endId`, if not only `id`.

`POST /api/annotations`

**Example Request**:

```json
POST /api/annotations HTTP/1.1
Accept: application/json
Content-Type: application/json

{
  "dashboardId":468,
  "panelId":1,
  "time":1507037197339,
  "isRegion":true,
  "timeEnd":1507180805056,
  "tags":["tag1","tag2"],
  "text":"Annotation Description"
}
```

**Example Response**:

```json
HTTP/1.1 200
Content-Type: application/json

{
    "message":"Annotation added",
    "id": 1,
    "endId": 2
}
```

## Create Annotation in Graphite format

Creates an annotation by using Graphite-compatible event format. The `when` and `data` fields are optional. If `when` is not specified then the current time will be used as annotation's timestamp. The `tags` field can also be in prior to Graphite `0.10.0`
format (string with multiple tags being separated by a space).

`POST /api/annotations/graphite`

**Example Request**:

```json
POST /api/annotations/graphite HTTP/1.1
Accept: application/json
Content-Type: application/json

{
  "what": "Event - deploy",
  "tags": ["deploy", "production"],
  "when": 1467844481,
  "data": "deploy of master branch happened at Wed Jul 6 22:34:41 UTC 2016"
}
```

**Example Response**:

```json
HTTP/1.1 200
Content-Type: application/json

{
    "message":"Graphite annotation added",
    "id": 1
}
```

## Create Annotation in Graphite format

Creates an annotation by using Graphite-compatible event format. The `when` and `data` fields are optional. If `when` is not specified then the current time will be used as annotation's timestamp. The `tags` field can also be in prior to Graphite `0.10.0`
format (string with multiple tags being separated by a space).

`POST /api/annotations/graphite`

**Example Request**:

```json
POST /api/annotations/graphite HTTP/1.1
Accept: application/json
Content-Type: application/json

{
  "what": "Event - deploy",
  "tags": ["deploy", "production"],
  "when": 1467844481,
  "data": "deploy of master branch happened at Wed Jul 6 22:34:41 UTC 2016"
}
```

**Example Response**:

```json
HTTP/1.1 200
Content-Type: application/json

{"message":"Graphite annotation added"}
```

## Update Annotation

`PUT /api/annotations/:id`

**Example Request**:

```json
PUT /api/annotations/1141 HTTP/1.1
Accept: application/json
Content-Type: application/json

{
  "time":1507037197339,
  "isRegion":true,
  "timeEnd":1507180805056,
  "text":"Annotation Description",
  "tags":["tag3","tag4","tag5"]
}
```

## Delete Annotation By Id

`DELETE /api/annotation/:id`

Deletes the annotation that matches the specified id.

**Example Request**:

```http
DELETE /api/annotation/1 HTTP/1.1
Accept: application/json
Content-Type: application/json
Authorization: Bearer eyJrIjoiT0tTcG1pUlY2RnVKZTFVaDFsNFZXdE9ZWmNrMkZYbk
```

**Example Response**:

```http
HTTP/1.1 200
Content-Type: application/json

{"message":"Annotation deleted"}
```

## Delete Annotation By RegionId

`DELETE /api/annotation/region/:id`

Deletes the annotation that matches the specified region id. A region is an annotation that covers a timerange and has a start and end time. In the Grafana database, this is a stored as two annotations connected by a region id.

**Example Request**:

```http
DELETE /api/annotation/region/1 HTTP/1.1
Accept: application/json
Content-Type: application/json
Authorization: Bearer eyJrIjoiT0tTcG1pUlY2RnVKZTFVaDFsNFZXdE9ZWmNrMkZYbk
```

**Example Response**:

```http
HTTP/1.1 200
Content-Type: application/json

{"message":"Annotation region deleted"}
```
