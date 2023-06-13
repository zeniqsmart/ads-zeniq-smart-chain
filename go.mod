module github.com/zeniqsmart/moeingads

go 1.18

require (
	github.com/coinexchain/randsrc v0.2.0
	github.com/dterei/gotsc v0.0.0-20160722215413-e78f872945c6
	github.com/minio/sha256-simd v0.1.1
	github.com/mmcloughlin/meow v0.0.0-20181112033425-871e50784daf
	github.com/stretchr/testify v1.8.3
)

require github.com/linxGnu/grocksdb v1.7.2

require (
	github.com/davecgh/go-spew v1.1.1 // indirect
	github.com/pmezard/go-difflib v1.0.0 // indirect
	golang.org/x/crypto v0.0.0-20191011191535-87dc89f01550 // indirect
	golang.org/x/sys v0.0.0-20190412213103-97732733099d // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
)

replace github.com/linxGnu/grocksdb => github.com/linxGnu/grocksdb v1.8.0
