# Key-Value Store API with Telegram Bot Integration

This project provides a key-value store API using Go and bbolt for database storage. Users can create and modify their API keys via a Telegram bot. Each user has their own separate database file.

## Cloud usage

1. Start telegram bot - https://t.me/kvestore_bot
2. Send /create_kv in bot

## Telegram Bot Commands


### `/docs`
Displays the documentation for all available commands to the user.

### `/create_kv`
Creates a new key-value (KV) store for the user. It generates a unique API key and creates a new BoltDB file to store the user's data. The API key is then sent back to the user.

### `/change_api_key`
Allows the user to change their existing API key. It generates a new API key, renames the BoltDB file with the new key, and sends the new API key to the user.

### `/view_bucket_keys`
Allows the user to view the keys stored in a specific bucket within their KV store. The user needs to provide the name of the bucket they want to view.

Usage: `/view_bucket BUCKET_NAME`

### `/list_buckets`
Lists all the buckets that the user has created in their KV store.

### `/download_kv`
Allows the user to download their entire KV store as a BoltDB file. The bot will send the file directly to the user.


## API Endpoints

#### Creating a new bucket

<details>
 <summary><code>PUT</code> <code><b>/</b><b>{bucketName}</b></code></summary>

##### Parameters

> | name      |  type     | data type   | description                 |
> |-----------|-----------|-------------|-----------------------------|
> | `bucketName` |  required | string      | Name of the bucket to create |

##### Responses

> | http code     | content-type         | response                              |
> |---------------|----------------------|---------------------------------------|
> | `200`         | `text/plain;charset=UTF-8` | `Bucket created successfully`          |
> | `405`         | `text/plain;charset=UTF-8` | `Bucket name 'system' not allowed`     |
> | `500`         | `text/plain;charset=UTF-8` | `Internal Server Error`                |

##### Example cURL

> ```shell
>  curl -X PUT -H "API-KEY: your_api_key" https://kvest.store/yourBucketName
> ```

</details>

#### Deleting an existing bucket

<details>
 <summary><code>DELETE</code> <code><b>/</b><b>{bucketName}</b></code></summary>

##### Parameters

> | name      |  type     | data type   | description                 |
> |-----------|-----------|-------------|-----------------------------|
> | `bucketName` |  required | string      | Name of the bucket to delete |

##### Responses

> | http code     | content-type            | response                              |
> |---------------|-------------------------|---------------------------------------|
> | `200`         | `text/plain;charset=UTF-8` | `Bucket deleted successfully`          |
> | `500`         | `text/plain;charset=UTF-8` | `Internal Server Error`                |

##### Example cURL

> ```shell
>  curl -X DELETE -H "API-KEY: your_api_key" https://kvest.store/yourBucketName
> ```

</details>

#### Creating/updating a key-value pair in a bucket

<details>
 <summary><code>PUT</code> <code><b>/</b><b>{bucketName}/{key}</b></code></summary>

##### Parameters

> | name      |  type     | data type   | description                 |
> |-----------|-----------|-------------|-----------------------------|
> | `bucketName` |  required | string      | Name of the bucket |
> | `key` |  required | string | Name of the key within the bucket |
> | None (body) |  required | object (JSON) | Value to be set for the key |

##### Responses

> | http code     | content-type            | response                              |
> |---------------|-------------------------|---------------------------------------|
> | `200`         | `text/plain;charset=UTF-8` | None                                   |
> | `400`         | `text/plain;charset=UTF-8` | `Bad Request`                          |
> | `500`         | `text/plain;charset=UTF-8` | `Internal Server Error`                |

##### Example cURL

> ```shell
>  curl -X PUT -H "API-KEY: your_api_key" -H "Content-Type: application/json" --data '{"key": "value"}' https://kvest.store/yourBucketName/yourKey
> ```

</details>

#### Retrieving a value for a key in a bucket

<details>
 <summary><code>GET</code> <code><b>/</b><b>{bucketName}/{key}</b></code></summary>

##### Parameters

> | name      |  type     | data type   | description                 |
> |-----------|-----------|-------------|-----------------------------|
> | `bucketName` |  required | string      | Name of the bucket |
> | `key` | required | string | Name of the key within the bucket |

##### Responses

> | http code     | content-type            | response                              |
> |---------------|-------------------------|---------------------------------------|
> | `200`         | `application/json`       | JSON object representing the value     |
> | `404`         | `text/plain;charset=UTF-8` | `Key not found`                        |
> | `500`         | `text/plain;charset=UTF-8` | `Internal Server Error`                |

##### Example cURL

> ```shell
>  curl -X GET -H "API-KEY: your_api_key" https://kvest.store/yourBucketName/yourKey
> ```

</details>

#### Deleting a key-value pair in a bucket

<details>
 <summary><code>DELETE</code> <code><b>/</b><b>{bucketName}/{key}</b></code></summary>

##### Parameters

> | name      |  type     | data type   | description                 |
> |-----------|-----------|-------------|-----------------------------|
> | `bucketName` |  required | string      | Name of the bucket |
> | `key` |  required | string | Name of the key within the bucket |

##### Responses

> | http code     | content-type            | response                              |
> |---------------|-------------------------|---------------------------------------|
> | `200`         | `text/plain;charset=UTF-8` | None                                   |
> | `500`         | `text/plain;charset=UTF-8` | `Internal Server Error`                |

##### Example cURL

> ```shell
>  curl -X DELETE -H "API-KEY: your_api_key" https://kvest.store/yourBucketName/yourKey
> ```

</details>

---

## Migrate on your server

> ```shell
> git clone https://github.com/Split174/kvest.git
> cd kvest
> docker-compose up -d
> ```
Download db file from bot `/download_db`.
> ```shell
> docker cp YOURKV.DB container_id:/data/YOURKV.DB
> ```

## Project Structure
```
├── api
│   ├── api.go
│   └── api_test.go
├── Caddyfile
├── docker-compose.yml
├── Dockerfile
├── go.mod
├── go.sum
├── main.go
├── README.md
└── telegram_bot
    └── telegram_bot.go
```
## Requirements

- Go 1.16 or later
- A valid Telegram Bot Token

## Dependencies

This project uses the following Go packages:
- [bbolt](https://github.com/etcd-io/bbolt)
- [gorilla/mux](https://github.com/gorilla/mux)
- [go-telegram-bot-api](https://github.com/go-telegram-bot-api/telegram-bot-api)

