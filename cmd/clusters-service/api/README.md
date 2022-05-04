# `cluster-services` API

The `cluster-services` HTTP API is generated from this here `cluster_services.proto` protobuf file.

See the [API Standardisation doc](https://gist.github.com/bigkevmcd/d97ddd38c5d82430bdc85f783e26b72e) for motivation.

## Get the tooling in place.

1. Install `buf`, [instructions](https://docs.buf.build/installation)
2. Install all the protobuf plugins `(cd cmd/clusters-service && make install)`

## How to add a new HTTP endpoint

1. Add a new `rpc` declaration to the `service` following the pattern there.

   ```protobuf
   rpc ListBananas(ListBananasRequest) returns (ListBananasResponse) {
       option (google.api.http) = {
       get : "/bananas"
       };
   }
   ```

2. Add the supporting Request and Response types

   ```protobuf
   message ListBananasRequest {
       // How many bananas do you want? (Will be used as a query string)
       int32 count = 1;
   }
   message ListBananasResponse {
       repeated string bananaTypes = 1;
       int32 total = 2;
   }
   ```

3. Run `make generate`
4. Add a new method to `./pkg/server/server.go` to actually return some bananas: (untested):
   ```golang
   func (s *server) ListBananas(ctx context.Context, msg *capiv1.ListBananasRequest) (*capiv1.ListBananasResponse, error) {
       if msg.count > 2 {
           return errors.New("We only have 2 types or bananas!")
       }
       bananaTypes := []*string{"green", "yellow"}
       return &capiv1.ListBananasResponse{BananaTypes: bananaTypes, Total: 2}, nil
   }
   ```
5. Git add and commit all new generated files!
