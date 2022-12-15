# gobot
Golang IRC bot


#### Build

Install deps
```
github.com/v0l/go-ircevent
github.com/ChimeraCoder/anaconda
github.com/robertkrimen/otto
golang.org/x/net/html
```

Compile
```
go build
```

Run
```
./gobot
```


## Twitter auth
Install deps:
```bash
composer install
composer dumpautoloader
mkdir tokens
```

Add API Key and Secret to `auth.php`

Run PHP dev server and open http://localhost:9999/auth.php

```
php -S localhost:9999 -t $(pwd)
```

Click on the Auth Link and authorize the app

Restart or run the bot