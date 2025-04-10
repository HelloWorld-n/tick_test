# API Documentation

Hosted at 127.0.0.1:4041 .

Account and Book endpoints require postgres installed and running.  
Put path to postgres at `./.config/dbPath.txt`.  

This document provides examples of requests and responses for the available API endpoints.  


---

## Index Endpoint

---

### GET `/v1`

Example Response:  
```json
{
  "Iteration": 0,
  "Now": "2025-03-07T19:50:40Z"
}
```
> Returns the current iteration count (starts at 0 and manipulated via Manipulator Endpoints) and the current UTC timestamp.

---

## Account Endpoints

---

### POST `/v1/accounts/register`

Example Request:
```json
{
  "Username": "exampleUser",
  "Password": "SecurePassword",
  "SamePassword": "SecurePassword",
  "Role": "User"
}
```
> Creates a new account. If the "Role" field is omitted, it defaults to "User".
> Returns the created account data on success or an error if the username already exists.

---

### GET `/v1/accounts/all`

Example Response:
```json
[
  {
    "Username": "ExampleUser",
    "Role": "User"
  },
  {
    "Username": "Librarian",
    "Role": "BookKeeper"
  }
]
```
> Retrieves a list of all user accounts.

---

### DELETE `/v1/accounts/delete`

> Deletes the account identified by the specified username.

---

### PATCH `/account/modify`

Example Request:
```json
{
  "Username": "newUsername",
  "Password": "new SecurePassword()",
  "SamePassword": "new SecurePassword()"
}
```
> Updates an existing account.
> Fields provided in the request will be updated (e.g. changing the username or password).
> Ensures that the new password meets the requirements and that "Password" and "SamePassword" match.

---

### PATCH `/v1/accounts/promote`

Example Request:
```json
{
  "Username": "user1",
  "Role": "Admin"
}
```
> Promotes an account by updating its role.
> The endpoint verifies that the account exists before applying the new role.
> Requires user with role `Admin`

---

## Book Endpoints

---

### GET `/v1/books/all`

Example Response:
```json
[
  {
    "Code": "GeneratedCode",
    "Title": "The Go Programming Language",
    "Author": "Alan Donovan"
  },
  {
    "Code": "DifferentCode",
    "Title": "Introducing Go",
    "Author": "Caleb Doxsey"
  }
]
```

> Retrieves a list of all books.

---

### GET `/v1/books/code/`*code*

Example Response:
```json
{
  "Code": "GeneratedCode",
  "Title": "The Go Programming Language",
  "Author": "Alan Donovan"
}
```

>  Retrieves a specific book by its unique code.

---

### POST `/v1/books`

Example Request:
```json
{
  "Title": "Learning Go",
  "Author": "Jon Bodner"
}
```

> Creates a new book entry.

---

### PATCH `/v1/books/`*code*

Example Request:
```json
{
  "Title": "Learning Go: Updated Edition",
  "Author": "Jon Bodner"
}
```

> Updates the title and/or author of the specified book. Only the provided fields will be updated.

---

### DELETE `/v1/books/`*code*

> Deletes the book with the specified code.

## Password Endpoints

---

### GET `/v1/password`
Example Response:
```json
["PASSWORD", "HASHES"]
```
> Retrieves the list of created passwords. (Only current session.)

---

### GET `/v1/password/rate/`*password*

Example Response:
```json
{
  "Password": "JTQ-29841-ε",
  "Score": 7
}
```
> Rates the given password by evaluating its length and character composition.

---

### POST `/v1/password/simple`

Example Request:
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

---

Example Request:
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

### POST `/v1/password/simple-stack`

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

### GET `/v1/sort`

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

### GET `/v1/sort/meta`

Example Response:
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

### GET `/v1/sort/code/`*code*

Example Response:
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

### POST `/v1/sort/increase`

Request Body:
```json
[3.14, 1.59, 2.65, 0.0]
```
> Sorts an array of numbers in increasing order.
> Use GET `/sort/code/`:code to see result.

---

### POST `/v1/sort/decrease`

Request Body:
```json
[3.14, 1.59, 2.65, 0.0]
```
> Sorts an array of numbers in decreasing order.
> Use GET `/sort/code/`*code* to see result.

---

### POST `/v1/sort/increase-abs`

Request Body:
```json
[-10, 5, -3, 7]
```
> Sorts an array of numbers by increasing absolute value.
> Use GET `/sort/code/`*code* to see result.

---

### POST `/v1/sort/decrease-abs`

Request Body:
```json
[-10, 5, -3, 7]
```
> Sorts an array of numbers by decreasing absolute value.
> Use GET `/sort/code/`*code* to see result.

---

### POST `/v1/sort/calculative/intensive`

Request Body:
```json
[[1, 2, 3], [3, 2, 1], [2, 2, 2]]
```
> Sorts an array of arrays based on the result of an intensive calculation performed on each sub-array.
> Use GET `/sort/code/`*code* to see result.

---

### POST `/v1/sort/calculative/calculate-once`

Request Body:
```json
[[1, 2, 3], [3, 2, 1], [4, 5, 6]]
```
> Sorts an array of arrays after performing a single calculation per element, ensuring that the intensive computation is executed only once for each.
> Use GET `/sort/code/`*code* to see result.

---

### DELETE `/v1/sort/calculative/delete-all`

> Deletes all sorting information.

---

## Manipulator Endpoints

---

### GET `/v1/manipulators`
Example Response:
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

### GET `/v1/manipulators/code/`*code*

Example: GET /manipulator/code/COMPLICATION
```json
{
  "Duration": "PT5S",
  "Value": 2
}
```
> Retrieves the configuration data for a specific iteration manipulator by its code.

---

### POST `/v1/manipulators`

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

### PATCH `/v1/manipulators/code/`*code*

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

### DELETE `/v1/manipulators/code/`*code*

> Deletes an iteration manipulator identified by its unique code. No request body is required.

---
