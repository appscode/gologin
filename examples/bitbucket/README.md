
# BitBucket Login

Login with BitBucket allows users to login to any web app with their BitBucket account.

## Web

Package `gologin` provides Go handlers for the BitBucket OAuth2 Authorization flow and for obtaining the BitBucket [User struct](https://github.com/dghubble/gologin/blob/master/bitbucket/verify.go).

### Getting Started

    go get github.com/dghubble/gologin/bitbucket
    cd $GOPATH/src/github.com/dghubble/gologin/examples/bitbucket
    go get .

## Example App

[main.go](main.go) shows an example web app which uses `gologin` to issue a client-side cookie session. For simplicity, no data is persisted.

Visit the BitBucket [Application Dashboard](https://developers.bitbucket.com/apps) to get you app's id and secret. Add `http://localhost:8080/bitbucket/callback` as a valid OAuth2 Redirect URL under Settings, Advanced.

<img src="https://storage.googleapis.com/dghubble/bitbucket-valid-callback.png">

Compile and run `main.go` from `examples/bitbucket`. Pass the client id and secret as arguments to the executable

    go run main.go -client-id=xx -client-secret=yy
    2015/09/25 23:09:13 Starting Server listening on localhost:8080

or set the `BITBUCKET_CLIENT_ID` and `BITBUCKET_CLIENT_SECRET` environment variables.

Here's what the flow looks like.

<img src="https://storage.googleapis.com/dghubble/bitbucket-web-login.gif">

### Authorization Flow

1. The "Login with BitBucket" link to the login handler directs the user to the BitBucket OAuth2 Auth URL to obtain a permission grant.
2. The redirection URI (callback handler) receives the OAuth2 callback, verifies the state parameter, and obtains a Token.
3. The success `ContextHandler` is called with a `Context` which contains the BitBucket Token and verified BitBucket User struct.
4. In this example, that User is read and used to issue a signed cookie session.

