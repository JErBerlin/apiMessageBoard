# Back's Message Board

## A message board with a public API for posting, and a private API to list and update

### Overview

This is a message board server exposing a RESTful API 

    1. used by unauthenticated user to post a new message

    2. used by an authenticated admin to view or edit a message

The application offers also a websockets communication to

    3. retrieve all the messages in streaming

### Flat-file database
Existing messages will be preloaded when the application starts, using a CSV file both as a source and as a flat-file 
database.

Here is a sample of the CSV file: https://drive.google.com/open?id=1sTfQWh58XTHgNy7fq04SPa79HZoTsjD 

This file (or a csv file with the same format and name: `messages.csv`) should be placed at the root of the directory 
from which the application or the docker container is run.

### Main features of the implementation and trade-offs

The design was kept as simple as possible, based around the csv file and sparing some kind of database other that this
flat-file system. A reason to do this would be the increased portability and maintainability. 

#### No proper database, no storage in memory: using indexation

Since the csv file containing the messages could be so large as a few GB, there is no storage in memory of the database 
as a whole. For the purpose of searching and editing a message by id, an indexing system is provided, which keeps the
start position (bytes) of the record in the file. This indexing is done often enough to ensure consistency, and this 
could have an impact on the performance. 

The sorting of data records is done also by means of these indexes, which don't store non-indexing information like the 
message text, so that the memory requirements are kept low. 

New messages are appended at the end of the file, provided with an automatically random generated id, that should not
(probabilistically) collide with pre-existing ids. 

Editing of a message is accomplished by duplicating the message (this means duplicating its id and cloning all the other
fields excepting the text), and its text field being updated with a new string (this is in fact the only editable field). 
The updated record is the appended at the end of the file and the old record is kept in its original position of the 
file. Even when it could be a concern to keep the old records, and in doing so to have records with duplicated id, the 
indexes take in account by design only the last occurrence of a given id and the data can keep its coherence. 

The decision of duplicating ids and keeping old records was made because the way in which read/write file operations work
in os systems, but also to ensure data coherence and make the system thread safe. As a result, the csv file can grow in 
size more than is needed, but we suppose the update operations to keep within reasonable limits. A point for keeping the 
old messages is to keep a log and to implement a kind of version system.

The records are always identified by a 16 hexadecimal number, that is usually given as a string the following format:

`A83087C2-562A-6904-FBA8-A3A7796E712B`

#### Mixed architecture API: RESTfull and straming capable

For the most part (posting, editing and viewing single messages) the API follows a RESTfull architecture. On the other 
hand it is hardly workable to send a large file following an http request to a RESTfull API. The present API uses 
websockets to send a large collection of JSON objects through streaming. 

### Accessing public and private API for posting, viewing and editing a single message

In the demonstration setting, the API will be listening and serving http at port 8080 of the localhost. The url paths 
for the different actions are the following:

1. for posting, make POST request at http://localhost:8080/new

You should specify a name, email, and a text for the message in JSON format as the body of the request, for instance:

`{
    "Name": "Jean Luc",
    "Email": "jean.luc.p@myship.sf",
    "Text": "new message from today"
}`

Since there is no formal data base, nor a query syntax, and the app is compiled statically, there are no big dangers in 
parsing external data. The only important data of a record that has to be correctly formatted is id, and they are 
produced automatically by a random generator of the app. For these reasons, there is no validation of the new message 
fields, other than a valid JSON formatting. 

The following actions must be preceded by basic authorization (admin/back-challenge)

2. for editing, make a POST request at http://localhost:8080/edit, specifying the id of the message and the new text of
the message in JSON format as the body of the request, for instance:

    `{
        "Id": "A83087C2-562A-6904-FBA8-A3A7796E712B",
        "Text": "today is yesterday"
    }`

3. for viewing a single message, make a GET request at http://localhost:8080/view/:id, for instance 

    `localhost:8080/view/9B3476D2-DBC0-0644-9656-FE2B98516981`

(Remember that this is only for demonstration purposes and that basic authorization wouldn't be used in any case in a 
production setting.)

### Web interface 

For demonstration purposes of the websocket streaming, an html page is served at the base-url:
 
`localhost:8080/`
 
 from which the list of all
records can be requested using the command 'all'. A single record can be viewed also by introducing its id in the 16-hex
string form. 

### Performance

For a sample file of about 250kB in size and keeping 500 records, the performance tested at localhost is obviously ideal. 
For much larger files (tested up to 1GB file with over 3 500 000 records) the performance becomes poor, because of the
limitations of the file based database and the simple indexation. This could be solved using enhanced file database 
systems, or even using a proper database. Implementing these extensions should be very straight forward. 

Another path to explore would be the introduction of concurrency for the search and indexing operations, but this could 
be a more sensible change, since the indexes should be made lockable and the write/read file operations should use other
libraries than the standard ones too, to make them thread safe.  

### Testing

There is no unit testing written for this app, as it is in development yet and a test driven design was not chosen. 
Before making the next step to production, tests should be written to sufficient extension.  

### To run the app through the Docker image

There is a provided Dockerfile to run the application in a docker container. 

To build a new image you can type from the directory where the dockerfile is:

`docker build -t back-message-board .`

To run the docker container:

`docker run -it back-message-board`

### Next steps

The next steps in this development should be

1. writing unit tests

2. improve the performance of the flat-file database, possibly by using a library like boltdb:

    `https://github.com/boltdb/bolt`

3. incorporate or improve concurrency in the write/read operations on the csv file, possibly by using operations of the 
log package, instead of the functions in the os package. 

4. coupled with concurrency in the read/write file operations, explore use of go routines for the indexing.

5. explore possibilities to bring together the RESTfull principles and the possibilities of streaming for the large 
files (a streaming API is not RESTfull). One possibility would be to download the messages.csv of a generated JSON file 
directly using other methods than an http request.