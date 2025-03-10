# API Documentation

Hosted at 127.0.0.1:4041 .

This document provides examples of requests and responses for the available API endpoints.


---

## Index Endpoint

---

### GET `/`

Example Response:
```json
{
  "Iteration": 0,
  "Now": "2025-03-07T19:50:40Z"
}
```
> Returns the current iteration count (starts at 0 and manipulated via Manipulator Endpoints) and the current UTC timestamp.

---

## Password Endpoints

---

### GET `/password`
Example Response:
```json
["PASSWORD", "HASHES"]
```
> Retrieves the list of created passwords. (Only current session.)

---

### GET `/password/rate/`*password*
Example Response:
```json
{
  "Password": "JTQ-29841-ε",
  "Score": 7
}
```
> Rates the given password by evaluating its length and character composition.

---

### POST `/password/simple`
Example Request
```json
{
  "Size": 10,
  "Charset": ["A", "B", "C", "D", "E"]
}
```
Example Response:
```json
"CABEABCEAA"
```
Example Request
```json
{
  "MinSize": 5,
  "MaxSize": 8,
  "Charset": ["A", "B", "C", "D", "E"]
}
```
Example Response:
```json
"EACEBDDD"
```
> Creates a simple password using a fixed size and a provided character set.  
> Note: When "Size" is provided, "MinSize" and "MaxSize" should not be provided.  

---

### POST `/password/simple-stack`

Example Request:
```json
[
  {
    "Size": 10,
    "Charset": ["A", "B", "C", "D"],
    "InclusionChances": 1.0
  },
  {
    "MinSize": 8,
    "MaxSize": 12,
    "Charset": ["x", "y", "z"],
    "InclusionChances": 0.5
  }
]
```
> Generates a composite password by stacking multiple password configurations. 
> Each configuration may use either a fixed "Size" or a range defined by "MinSize" and "MaxSize", along with an "InclusionChances" factor.

---

## Sort Endpoints

---

### GET `/sort`
Example Response:
```json
[
  {
    "Code": "RandomCode",
    "SortType": "increase",
    "StartedAt": "2025-03-07T12:00:00Z",
    "CompletedAt": "2025-03-07T12:00:01Z",
    "TimeTaken": "PT1S",
    "Result": [1, 2, 3]
  }
]
```
> Retrieves all stored sorting results.

---

### GET `/sort/meta`
```json
[
  {
    "SortType": "increase",
    "AverageTimeTaken": "PT0.5S",
    "MinTimeTaken": "PT0.2S",
    "MaxTimeTaken": "PT0.8S",
    "SampleSize": 5
  }
]
```
> Retrieves metadata summarizing sort performance for each sort type.

---

### GET `/sort/code/`*code*

Example Response
```json
{
  "Code": "RandomCode",
  "SortType": "increase",
  "StartedAt": "2025-03-07T12:00:00Z",
  "CompletedAt": "2025-03-07T12:00:01Z",
  "TimeTaken": "PT1S",
  "Result": [1, 2, 3]
}
```
> Retrieves a specific sorting result by its unique code given when requesting sorts.

---

### POST `/sort/increase`

Request Body:
```json
[3.14, 1.59, 2.65, 0.0]
```
> Sorts an array of numbers in increasing order.
> Use GET `/sort/code/`:code to see result.

---

### POST `/sort/decrease`

Request Body:
```json
[3.14, 1.59, 2.65, 0.0]
```
> Sorts an array of numbers in decreasing order.
> Use GET `/sort/code/`*code* to see result.

---

### POST `/sort/increase-abs`

Request Body:
```json
[-10, 5, -3, 7]
```
> Sorts an array of numbers by increasing absolute value.
> Use GET `/sort/code/`*code* to see result.

---

### POST `/sort/decrease-abs`

Request Body:
```json
[-10, 5, -3, 7]
```
> Sorts an array of numbers by decreasing absolute value.
> Use GET `/sort/code/`*code* to see result.

---

### POST `/sort/calculative/intensive`

Request Body:
```json
[[1, 2, 3], [3, 2, 1], [2, 2, 2]]
```
> Sorts an array of arrays based on the result of an intensive calculation performed on each sub-array.
> Use GET `/sort/code/`*code* to see result.

---

### POST `/sort/calculative/calculate-once`

Request Body:
```json
[[1, 2, 3], [3, 2, 1], [4, 5, 6]]
```
> Sorts an array of arrays after performing a single calculation per element, ensuring that the intensive computation is executed only once for each.
> Use GET `/sort/code/`*code* to see result.

---

### DELETE `/sort/calculative/delete-all`

> Deletes all sorting information.

---

## Manipulator Endpoints

---

### GET `/manipulator`
Example Response
```json
[
  {
    "Code": "COMPLICATION",
    "Data": {
      "Duration": "PT5S",
      "Value": 2
    }
  }
]
```
> Retrieves all iteration manipulators along with their configuration data.

---

### GET `/manipulator/code/`*code*

Example: GET /manipulator/code/COMPLICATION
```json
{
  "Duration": "PT5S",
  "Value": 2
}
```
> Retrieves the configuration data for a specific iteration manipulator by its code.

---

### POST `/manipulator`

Request Body:
```json
{
  "Duration": "PT5S",
  "Value": 2
}
```

> Creates a new iteration manipulator with the specified duration and value.
> Starts timer; after each Duration will manipulate Iteration count by Value.

---

### PATCH `/manipulator/code/`*code*

Request Body:
```json
{
  "Duration": "PT10S"
}
```
> Updates an existing iteration manipulator’s configuration. 
> Only the provided fields will be updated.
> If changing duration: current timer will cancel (without manipulation), new timer will start. 

---

### DELETE `/manipulator/code/`*code*

> Deletes an iteration manipulator identified by its unique code. No request body is required.

---
