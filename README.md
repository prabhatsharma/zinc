[![Go Report Card](https://goreportcard.com/badge/github.com/zinclabs/zinc)](https://goreportcard.com/report/github.com/zinclabs/zinc)
<!-- ![](https://codebuild.us-west-2.amazonaws.com/badges?uuid=eyJlbmNyeXB0ZWREYXRhIjoid3creXFaUHdOZmJvWWFXM0RZckJqV0xhZG4vT1ZWUkREK05oZExmT3JMMitGNGJwOHVIdCtKdjNEQzVqWXpLQlY1QjF2QXRIa0dIRjUvTzBsTE9LR0c0PSIsIml2UGFyYW1ldGVyU3BlYyI6IjRzdjc5bWZxU2hJYXNYNzciLCJtYXRlcmlhbFNldFNlcmlhbCI6MX0%3D&branch=main) -->

# Zinc Search Engine

Zinc is a search engine that does full text indexing. It is a lightweight alternative to Elasticsearch and runs using a fraction of the resources. It uses [bluge](https://github.com/blugelabs/bluge) as the underlying indexing library.

It is very simple and easy to operate as opposed to Elasticsearch which requires a couple dozen knobs to understand and tune which you can get up and running in 2 minutes

It is a drop-in replacement for Elasticsearch if you are just ingesting data using APIs and searching using kibana (Kibana is not supported with zinc. Zinc provides its own UI).

Check the below video for a quick demo of Zinc.

[![Zinc Youtube](./screenshots/zinc-youtube.jpg)](https://www.youtube.com/watch?v=aZXtuVjt1ow)

# Playground Server

You could try ZincSearch without installing using below details: 

|          |                                        |
-----------|-----------------------------------------
| Server   | https://playground.dev.zincsearch.com  |
| User ID  | admin                                  |
| Password | Complexpass#123                        |

Note: Do not store sensitive data on this server as its available to everyone on internet. Data will also be cleaned on this server regularly.

# Join slack channel

[![Slack](./screenshots/slack.png)](https://join.slack.com/t/zinc-nvh4832/shared_invite/zt-11r96hv2b-UwxUILuSJ1duzl_6mhJwVg)

# Why zinc

  While Elasticsearch is a very good product, it is complex and requires lots of resources and is more than a decade old. I built Zinc so it becomes easier for folks to use full text search indexing without doing a lot of work.

# Features:

1. Provides full text indexing capability
2. Single binary for installation and running. Binaries available under releases for multiple platforms.
3. Web UI for querying data written in Vue
4. Compatibility with Elasticsearch APIs for ingestion of data (single record and bulk API)
5. Out of the box authentication
6. Schema less - No need to define schema upfront and different documents in the same index can have different fields.
7. Index storage in disk (default), s3 or minio (experimental)
8. aggregation support

# Roadmap items:

Public roadmap is available at https://github.com/orgs/zinclabs/projects/3/views/1

Please create an issue if you would like something to be added to the roadmap.

# Screenshots

## Search screen
![Search screen](./screenshots/search_screen.jpg)

## User management screen
![Users screen](./screenshots/users_screen.jpg)

# Getting started


## Download / Installation / Run

Check installation [installation docs](https://docs.zincsearch.com/04_installation/)


## Data ingestion

### Single record

Check [single record ingestion docs](https://docs.zincsearch.com/ingestion/single-record/)

### Bulk ingestion

Check [bulk ingestion docs](https://docs.zincsearch.com/ingestion/bulk-ingestion/#bulk-ingestion)

### Fluent bit

Check [fluent-bit ingestion docs](https://docs.zincsearch.com/ingestion/fluent-bit/)

### Fluentd

Check [fluentd ingestion docs](https://docs.zincsearch.com/ingestion/fluentd/)

### Syslog-ng

Check [syslog-ng ingestion docs](https://docs.zincsearch.com/ingestion/syslog-ng/)

## API Reference

Check [Zinc API docs](https://docs.zincsearch.com/API%20Reference/)

# Releases

ZincSearch currently has most of its API contracts frozen. It's data format may still experience changes as we improve things. Currently ZincSearch is in beta. Data format should become highly stable when we move to GA (version 1).

# How to develop and contribute to Zinc

Check the [contributing guide](./CONTRIBUTING.md) . Also check the [roadmap items](https://github.com/orgs/zinclabs/projects/3)

# Who uses Zinc (Known users)?

1. [Quadrantsec](https://quadrantsec.com/)
1. [Paopao](https://github.com/rocboss/paopao-ce)

Please do raise a PR adding your details if you are using Zinc.



