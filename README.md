#
An automatic way of creating command line options to override fields from a struct.

## Installation
`go get -u github.com/uber-go/flagoverride`


## Overview
Typically, if one wants to load from a config file (e.g. yaml), one has to
define a proper struct, then load values into it (e.g. yaml.Unmarshal()).
However, there are situations where we want to load most of the configs from
the file and to override some of the configs.

Let's say we use a yaml to config our Db connections and upon start of the
application we load from the yaml file to get the necessary parameters to
create the connection. Our base.yaml looks like this:

```yaml
  base.yaml
  ---
  mysql:
    user: 'foo'
    password: 'xxxxxx'
    mysql_defaults_file: ./mysql_defaults.ini
    mysql_socket_path: /var/run/mysqld/mysqld.sock
    ... more config options ...
```

we want to load all the configs from it but we want to provide some
flexibility for the program to connect via a different db user. We could
define a --user command flag then after loading the yaml file, we override
the user field with what we get from --user flag.

If there are many overriding like this, manual define these flags is
tedious. This package provides an automatic way to define this override,
which is, given a struct, it'll create all the flags which are name using
the field names of the struct. If one of these flags are set via command
line, the struct will be modified in-place to reflect the value from command
line, therefore the values of the fields in the struct are overridden.

YAML is just used as an example here. In practice, one can use any struct tdefine flags.

Let's say we have our configuration object as the following:

```go
  type logging struct {
  	 Interval int
  	 Path     string
  }

  type socket struct {
  	 ReadTimeout  time.Duration
  	 WriteTimeout time.Duration
  }

  type tcp struct {
  	 ReadTimeout time.Duration
  	 socket
  }

  type network struct {
  	 ReadTimeout  time.Duration
  	 WriteTimeout time.Duration
  	 tcp
  }

  type Cfg struct {
  	 logging
  	 network
  }
```

The following code:

```go
  func main() {
    c := &Cfg{}
    flags.ParseArgs(c, os.Args[1:])
  }
```

will create the following flags:

```
  -logging.interval int
        logging.interval
  -logging.path string
        logging.path
  -network.readtimeout duration
        network.readtimeout
  -network.tcp.readtimeout duration
        network.tcp.readtimeout
  -network.tcp.socket.readtimeout duration
        network.tcp.socket.readtimeout
  -network.tcp.socket.writetimeout duration
        network.tcp.socket.writetimeout
  -network.writetimeout duration
        network.writetimeout
```

flags to subcommands are naturally supported.

```go
  func main() {
    cmd := os.Args[1]
    switch cmd {
      case "new"
      c1 := &Cfg1{}
      ParseArgs(c1, os.Args[2:])
    case "update":
      c2 := &Cfg2{}
      ParseArgs(c2, os.Args[2:])

    ... more sub commands ...
    }
  }
```

One can set Flatten to true when calling `NewFlagMakerAdv`, in which case,
flags are created without namespacing. For example,

```go
  type auth struct {
   Token string
   Tag   float64
  }

  type credentials struct {
   User     string
   Password string
   auth
  }

  type database struct {
   DBName    string
   TableName string
   credentials
  }

  type Cfg struct {
   logging
   database
  }

  func main() {
   c := &Cfg{}
   flags.ParseArgs(c, os.Args[1:])
  }
```

will create the following flags:

```
  -dbname string
        dbname
  -interval int
        interval
  -password string
        password
  -path string
        path
  -tablename string
        tablename
  -tag float
        tag
  -token string
        token
  -user string
        user
```

Please be aware that usual GoLang flag creation rules apply, i.e., if there are
duplication in flag names (in the flattened case it's more likely to happen
unless the caller make due diligence to create the struct properly), it panics.  

Note that not all types can have command line flags created for.  

`map`, `channel` and function type will not define a flag corresponding to the field.  

Pointer types are properly handled and slice type will create multi-value command line flags.  

That is, e.g. if a field foo's type is `[]int`, one can use
--foo 10 --foo 15 --foo 20 to override this field value to be
`[]int{10, 15, 20}`. For now, only `[]int`, `[]string` and `[]float64` are supported in this fashion.  

<hr>
Released under the [MIT License](LICENSE.txt).
