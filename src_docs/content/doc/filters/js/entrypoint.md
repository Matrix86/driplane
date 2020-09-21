---
weight: 2
title: "Entrypoint"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Entrypoint

The JS file consists of one or more functions, and they can be used in different filters or in combination between them.
The function must respect some constraints:

 * the function's name contained in the `function` parameter of the filter has to start with a capital letter;
 * the function prototype must 3 variables;
 * the function must return a JS object with at least the `filtered` field;

### Function's prototype

The function's name specified in the `function` parameter of the filter, receives 3 input parameters:
 * main: it is a string with the content of the field `main` of the input Message;
 * extra: a JS object that can be seen like an associative array, containing the extra fields of the input Message;
 * params: a JS object like the previous, but it contains the configurations from the `custom` and the `general` sections. 

So for example we can define a function like the follow:

```javascript
function Entry(main, extra, params) {
 ...
}
```

### Return value

The JS function has to return a value back to the filter, so that the filter can use it to propagate the Message or drop it.
The method used on `function` parameter can return a JS object containing at least the `filtered` field.
If this field has been set to true, the `Message` will be sent to the next filter, otherwise if the `filtered` field has been set to false, the filter will drop the Message.

We would change the fields of the `Message`, and to do that we can use the `data` field in the returned object.
It should be an associative array and it will be mapped in a map[string]string object in the Go env. 
The key of the array's row is the name of the field to add or change, while the value is the string that field's Message will contain after the return.

```javascript
function Entry(mainData, extra, params) {
    return {
        "filtered": true,
        "data": {
            "main": "main field of the input Message changed",
            "new_field": "new_field will be added as Message's extra field"
        }
    };
}
```