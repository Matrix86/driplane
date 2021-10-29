---
weight: 2
title: "Plugin Packages"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Packages

Since the JS doesn't have all the useful functions to perform operations like on file manipulation or http request (see [otto docs](https://godoc.org/github.com/robertkrimen/otto) for more info), `driplane` maps and export to the JS environment some Go functions as packages.


{{< alert theme="warning" >}}
These mapped functions are not definitive and they could change over time.
{{< /alert >}} 

### File

This package contains functions for manipulating files.

#### Return

The return value will be a JS object containing 2 fields:

| Name | Type | Description |
| --- | --- | --- |
| Status | _BOOL_ | if true the operation was successful |
| Error | _STRING_ | if status is false it contains the reason of the failure |

#### Functions

| Prototype | Description |
| --- | --- |
| file.Move(src string, dest string) | Move the file `src` to the new position `dest` |
| file.Copy(src string, dest string) | Copy the file `src` to the position `dest` |
| file.Truncate(filename string, size int) | set or adjust the file called `filename` by `size` bytes |
| file.Delete(filename string) | Remove the file called `filename` |
| file.Exists(filename string) | Return Status = true if `filename` exists on disk |
| file.AppendString(filename string, text string) | Append the string `text` to the file `filename` |

### Log

This package contains functions for writing strings on the logs.

#### Return

It doesn't have a return value.

#### Functions

| Prototype | Description |
| --- | --- |
| log.Info(format string, ...) | Write an Info line on the log |
| log.Error(src string, ...) | Write an Error line on the log |
| log.Debug(src string, ...) | Write a Debug line on the log |

### Strings

This package contains functions for manipulating strings.

#### Return

The return value will be a JS object containing 2 fields:

| Name | Type | Description |
| --- | --- | --- |
| Status | _BOOL_ | if true the operation was successful |
| Error | _STRING_ | if status is false it contains the reason of the failure |

#### Functions

| Prototype | Description |
| --- | --- |
| strings.StartsWith(str string, substr string) | Status = true if `str` start with `substr` |

### Util

This package contains miscellaneous functions.

#### Return

The return value will be a JS object containing 2 fields:

| Name | Type | Description |
| --- | --- | --- |
| Status | _BOOL_ | if true the operation was successful |
| Error | _STRING_ | if status is false it contains the reason of the failure |
| Value | _STRING_ | this string contains the returned value |

#### Functions

| Prototype | Description |
| --- | --- |
| util.Sleep(seconds int) | wait for `seconds` seconds |
| util.Getenv(name string) | get the value of the `name` environment variables, if it exists, and return it on the `Value` field of the returned value |
| util.Md5File(filename string) | calculate the MD5 hash of the `filename` and return it on the `Value` field |

### Http

This package contains functions to perform HTTP requests.

#### Return

The return value will be a JS object containing 4 fields:

| Name | Type | Description |
| --- | --- | --- |
| Status | _BOOL_ | if true the operation was successful |
| Error | _STRING_ | if status is false it contains the reason of the failure |
| Response | _OBJECT_ | it returns the [http.Response](https://golang.org/pkg/net/http/#Response) object |
| Body | _STRING_ | it contains the body of the response converted in string |

#### Functions

| Prototype | Description |
| --- | --- |
| http.Request(method string, uri string, headers interface{}, data interface{}) |  |
| http.Get(url string, headers map[string]string) |  |
| http.Post(url string, headers map[string]string, data interface{}) |  |
| http.DownloadFile(filepath string, method string, uri string, headers interface{}, data interface{}) |  |
| http.UploadFile(filename string, fieldname string, method string, uri string, headers interface{}, data interface{}) |  |

### Cache

This package contains functions for add and get values from the **global** cache.

#### Return

The return value will be a JS object containing the following fields:

| Name | Type | Description |
| --- | --- | --- |
| Status | _BOOL_ | if true the operation was successful |
| Error | _STRING_ | if status is false it contains the reason of the failure |
| Value | _STRING_ | if status is true it contains the resulting value |

#### Functions

| Prototype | Description |
| --- | --- |
| cache.Put(key string, value string, ttl int64) | add the `value` in the cache using the key `key` and it will be deleted after `ttl` seconds |
| cache.Get(key string) | get the value stored in the cache with the key `key` |
