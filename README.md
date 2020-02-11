# Back's Message Board

## A message board with a public API for posting, and a private API to list and update

### Overview

Create the backend of a message board application. It exposes 2 RESTful APIs: 

1. public API used by unauthenticated user to post a new message 
2. private API used by an authenticated administrator to list all existing message and update an existing message's text.

### Pre-loading messages

Existing messages will be preloaded when the application starts using a CSV file. 

Here is a sample of the CSV file: https://drive.google.com/open?id=1sTfQWh58XTHgNy7fq04SPa79HZoTsjD 

Note: this file is only a sample and that its size is unknown and could be as big as a few gigabytes.

### Public API

This API do not require any authentication to the accessed.
It exposes only one endpoint to create a new message.

### Private API

This API requires an authentication using login/password (admin/back-challenge)

It exposes 3 endpoints: 

1. list all messages, ordered anti-chronologically

2. view a specific message identified by its id

3. update the text of a specific message identified by its id

### Evaluation criteria 

The code is readable, well commented and maintainable. Pick the language of your choice. We prefer Golang but we are comfortable with Python and Node.js as well. The project contains some tests. The project contains a README.md with the API documentation.
