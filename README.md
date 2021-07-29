# Miniproject - Key Value Store Server
The following project implements a distributed key-value store server that can handle concurrent requests from clients while maintaining a memory usage of less than 128MB.

## Contributor
- Shan Liu https://github.com/Shanni

## Building and Dependencies
This project makes use of the `proto3` library to encode/decode all messages passed between the server and the clients. All clients must have their protobuf source also generated with `proto3`, more specifically version 3.14.0 (though the exact version should not matter).

If you need to modify the protobuf protocal edit `*.proto` files and regenerate
the go code with: `protoc --go_out=. proto/* --experimental_allow_proto3_optional`

To run a single server, enter the following command in the root of this repository:
```
$go build -o dht-server

$docker build -t dht .

$docker run dht ./dht-server 3331 servers.txt
```

The hashcode is a integer in 0.255 that represents the keyspace that that server
is in charge of.

## Testing
There exists a test client in the `test-client` folder that implements some additional tests on top of the provided test client from the instructor. The tests included in this client are:
1. PUT commands that trigger the INVALID_KEY_ERR and INVALID_VAL_ERR errors.
2. GET and REMOVE commands that trigger the KEY_DNE_ERR error.
3. IS_ALIVE, PID, and GET_MEMBERSHIP_CNT commands.

To run the test client, enter the following command in the root of this repository:
```
go run test-client/testclient.go [SERVER IP ADDRESS] [SERVER PORT NUMBER]
```

## Code Contributing Guidelines
* Any commits to own branches, name as you like
* Any updates to main or milestone# *only through pull request with approves*
* *only merge* no squash, because actual commit history is required for grading
* Make sure at least one code review is given before merging

## Design Process Guidelines
* Document ideas before proceeding to implementation so we discuss them together
* Use this [Design Document](https://docs.google.com/document/d/1OL3UIhUURG6v-UgW2-vABsq4bqgdIP5T1kH4QY9Bgm8/edit)

MSG type:
    0: request
    1: response
    2: ack -> empty payload.