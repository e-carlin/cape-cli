# cli

Cape CLI

## Overview

### Install

Cape CLI can be simply installed with:

```
go install github.com/capeprivacy/cli/cmd/cape
```

Note: Make sure your $HOME/go/bin directory is in your $PATH.

### Cape Login

Log into Cape by running `cape login`:

```
$ cape login
Your CLI confirmation code is: <RANDOM CODE>
Visit this URL to complete the login process: https://maestro-dev.us.auth0.com/activate?user_code=<RANDOM CODE>
```

If your terminal is able to it will auto-launch a browser. Finish the log in process and confirm that the code matches
the code you are seeing in the browser. If your terminal can't launch a browser you can manually visit the link and complete
the process that way.

### Run Simple Test Function

While developing the function you want to run in Cape you can simple use the `cape test` command
to test against an actual enclave. In the example below there is a file containing the input data and a
directory containing the function to be run.

```
$ ls
input_data test_func

$ cape test test_func input_data
Success! Results from your function:
<RESULTS GO HERE>
```

Any logging output is output to stderr while results are output to stdout.

See `cape test --help` for more options.

### Deploy Function

Once your function is finalized you can deploy it to Cape for future use using `cape deploy`. The set up is similar
to `cape test`.

```
$ ls
test_func

$ cape deploy test_func
Success! Deployed function to Cape\nFunction ID ➜ <FUNCTION ID>\n
```

`<FUNCTION ID>` is a UUID that will then be used to pass to `cape run`.

See `cape deploy --help` for more options.

### Run Function

You and other users can use your deployed function by running the `cape run` command:

```
$ ls
input_data

$ cape run <FUNCTION ID> input_data
Success! Results from your function:
<RESULTS GO HERE>
```

Any tracing output is output to stderr while results are output to stdout.

See `cape run --help` for more options.

## Build

```
go build ./cmd/cape
```

## Config

For login purposes the following environment variables can be configured:

```
CLI_HOSTNAME                String    https://maestro-dev.us.auth0.com
CLI_CLIENT_ID               String    yQnobkOr1pvdDAyXwNojkNV2IPbNfXxx
CLI_AUDIENCE                String    https://newdemo.capeprivacy.com/v1/
CLI_LOCAL_AUTH_DIR          String    .cape
CLI_LOCAL_AUTH_FILE_NAME    String    auth
```
