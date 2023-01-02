module gitlab.com/alielgamal/hfid/example/redis

go 1.19

require (
	github.com/alicebob/miniredis/v2 v2.23.1
	github.com/go-redis/redis/v8 v8.11.5
	gitlab.com/alielgamal/hfid v0.0.0-20230102072626-b1929db24d94
	gitlab.com/alielgamal/hfid/redis v0.0.0-20230102072626-b1929db24d94
)

require (
	github.com/alicebob/gopher-json v0.0.0-20200520072559-a9ecdc9d1d3a // indirect
	github.com/cespare/xxhash/v2 v2.2.0 // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/yuin/gopher-lua v0.0.0-20220504180219-658193537a64 // indirect
)

replace gitlab.com/alielgamal/hfid => ../..

replace gitlab.com/alielgamal/hfid/redis => ../../redis
