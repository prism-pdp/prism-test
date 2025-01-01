# prism-test

# Update prism-go

1. Run docker container of harness
```
make harness/shell
```

2. Remove github.com/prism/prism-go from go.mod and go.sum

3. Download new prism-go
```
go get github.com/prism/prism-go
```

4. Exit from the container

5. Build docker image of harness
```
make build-img
```

# Update prism-sol

1. Move to prism-sol directory
```
cd testnet/prism-sol
```

2. Update source codes with git pull command