Simple console client to work with links manager server https://github.com/viktor-br/links-manager

Commands

Create new link:
```
cmd> http://google.com #google #search powerful search server
```

Create new user:
```
cmd> ua
```

Authenticate on remote server (receiving token with using saved credentials):
```
cmd> auth
```

Change saved credentials:
```
cmd> credentials
```

Exit:
```
cmd> exit
```

TODO

1. Token reading and receiving from remote should support concurrent access.
2. Buffer and run in parallel CRUD for items and CRUD for users.
3. Tests.
4. Help command.
5. Save and read configuration from file.
