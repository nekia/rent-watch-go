#!/bin/bash

IP=$(hostname -I)

# docker run --rm -i -v $(pwd)/../../protobuf/checker:/protos --rm fullstorydev/grpcurl -plaintext -d @ -import-path /protos -proto checker.proto ${IP[0]% }:50051 checker.Checker/CheckByUrl < ./test-input-CheckByUrl.json

docker run --rm -i -v $(pwd)/../../protobuf/checker:/protos --rm fullstorydev/grpcurl -plaintext -d @ -import-path /protos -proto checker.proto ${IP[0]% }:50051 checker.Checker/CheckByRoomDetail < ./test-input-CheckByRoomDetail.json

# docker run --rm -i -v $(pwd)/../../protobuf/checker:/protos --rm fullstorydev/grpcurl -plaintext -d @ -import-path /protos -proto checker.proto ${IP[0]% }:50051 checker.Checker/UpdateCheckStatus < ./test-input-UpdateCheckStatus.json

