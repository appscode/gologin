
# Slack Login

Login with Slack allows users to login to any web app with their Slack account.

## Web

Package `gologin` provides Go handlers for the Slack OAuth2 Authorization flow and for obtaining the Slack [User struct](https://github.com/dghubble/gologin/blob/master/slack/verify.go).

### Getting Started

    go get github.com/dghubble/gologin/slack
    cd $GOPATH/src/github.com/dghubble/gologin/examples/slack
    go get .

## Example App

[main.go](main.go) shows an example web app which uses `gologin` to issue a client-side cookie session. For simplicity, no data is persisted.

Visit the Slack [Application Dashboard](https://developers.slack.com/apps) to get you app's id and secret. Add `http://localhost:8080/slack/callback` as a valid OAuth2 Redirect URL under Settings, Advanced.

<img src="https://storage.googleapis.com/dghubble/slack-valid-callback.png">

Compile and run `main.go` from `examples/slack`. Pass the client id and secret as arguments to the executable

    go run main.go -client-id=xx -client-secret=yy
    2015/09/25 23:09:13 Starting Server listening on localhost:8080

or set the `SLACK_CLIENT_ID` and `SLACK_CLIENT_SECRET` environment variables.

Here's what the flow looks like.

<img src="https://storage.googleapis.com/dghubble/slack-web-login.gif">

### Authorization Flow

1. The "Login with Slack" link to the login handler directs the user to the Slack OAuth2 Auth URL to obtain a permission grant.
2. The redirection URI (callback handler) receives the OAuth2 callback, verifies the state parameter, and obtains a Token.
3. The success `http.Handler` is called with a `Context` which contains the Slack Token and verified Slack User struct.
4. In this example, that User is read and used to issue a signed cookie session.

