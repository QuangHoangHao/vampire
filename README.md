# Vampire Golang Kit

### Installation: Require Go 1.19
```shell
go install github.com/QuangHoangHao/vampire@latest
```

#### Guild:
```shell
vampire -h
```
Output:

```Plain Text
DI          Create DI
completion  Generate the autocompletion script for the specified shell
controller  Create Controller
help        Help about any command
init        Init Project
repo        Create Repo
service     Create Service
worker      Create Worker
   ```

Example : Init Project 
```shell
vampire init
```

With Answer:
```shell
? Module name: greet
? Generate Dockerfile? Yes
? Choose type: API
```

The generated files look like:

```Plain Text
   ├── greet
   │   ├── cmd
   │   │   └── main.go
   │   ├── .gitignore
   │   └── development.env
   │       Dockerfile 
   │       go.mod
   │       go.sum
   │       production.env
   │       README.md
   └
   ```
