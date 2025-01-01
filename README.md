# dpduado-test

# Update dpduado-go

1. Run docker container of harness
```
make harness/shell
```

2. Remove github.com/dpduado/dpduado-go from go.mod and go.sum

3. Download new dpduado-go
```
go get github.com/dpduado/dpduado-go
```

4. Exit from the container

5. Build docker image of harness
```
make build-img
```

# Update dpduado-sol

1. Move to dpduado-sol directory
```
cd testnet/dpduado-sol
```

2. Update source codes with git pull command