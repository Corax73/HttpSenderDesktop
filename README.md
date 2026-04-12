## HttpSenderDesktop

Simple desktop application, no registration or data storage, with a graphical interface for sending http requests. The main goal is local testing of developed applications.

Libraries used https://github.com/fyne-io/fyne, https://github.com/fyne-io/fyne-cross (for testing), https://github.com/golang-design/clipboard and others from the go standard library.

### Explanations:
1. The query string is sent as is, with GET parameters.
2. JSON with headers.
3. JSON with parameters.
4. Delay between requests (with a given repetition).
5. Number of request repetitions.
6. Menu with selection of request method. There is no default value.
7. Setting up a basic authentication login and password.
8. Setting cookies: name, value, expiration date (in hours).
9. Outputting the answer. When repeated, data is added.
10. Save state (entered request data) under a specific name.
11. Load one of the previously saved states.

The buttons also have self-explanatory names.

<p align="cen-er"><img src="readmeImg/screen.jpg"></p>
