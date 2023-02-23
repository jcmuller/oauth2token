# oauth2token

Retrieve an oauth2 authentication code from a google app and exchange it for a token.

## Installation

```
$ go install github.com/jcmuller/oauth2token@latest
go: downloading github.com/jcmuller/oauth2token vVERSION
```

## Configuration

It depends on having a valid configuration file in `~/.config/oauth2token/config.json`. An
example is:

```json
{
  "web": {
    "client_id": "SOME_CLIENT_ID",
    "project_id": "PROJECT_ID",
    "auth_uri": "https://accounts.google.com/o/oauth2/auth",
    "token_uri": "https://oauth2.googleapis.com/token",
    "auth_provider_x509_cert_url": "https://www.googleapis.com/oauth2/v1/certs",
    "client_secret": "CLIENT_SECRET",
    "redirect_uris": [
      "http://localhost:54321/"
    ]
  }
}
```

and the required scopes in `~/.config/oauth2token/scopes.json`. A good example is

```json
["https://mail.google.com/"]
```

The callback server will be started listening on the first redirect URI's port.
In this example, it would be `:54321`.

## Usage

### First run:

```
$ oauth2token
Please open this URL to authorize access: https://accounts.google.com/o/oauth2/auth?access_type=offline&client_id=CLIENT_ID&redirect_uri=http%3A%2F%2Flocalhost%3A54321%2F&response_type=code&scope=https%3A%2F%2Fmail.google.com%2F&state=RANDOM_STATE
A_VALID_TOKEN
```

### Successive runs:

```
$ oauth2token
A_VALID_TOKEN
```
