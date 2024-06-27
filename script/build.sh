CGO_ENABLED=0
GOOS=$(go env GOOS)
GOARCH=amd64
BUILD_VERSION=${BUILD_VERSION:=0.0.0}
COMMIT_HASH=${COMMIT_HASH:=dev}
BUILD_TIME=${BUILD_TIME:=dev}
BRANCH=${BRANCH:=dev}
ROOT=$( cd "$( dirname "$0" )/.." && pwd )
BUILD_FOLDER=target
PROJECT_NAME=go-todolist-grpc

mkdir -p $ROOT/$BUILD_FOLDER
cp -r $ROOT/app.env $ROOT/$BUILD_FOLDER/
cd $ROOT/cmd/$PROJECT_NAME
go mod tidy
go build -ldflags "-X 'main.buildVersion=$BUILD_VERSION' -X 'main.commitHash=$COMMIT_HASH' -X 'main.buildTime=$BUILD_TIME' -X 'main.branch=$BRANCH'" -o $ROOT/$BUILD_FOLDER/$PROJECT_NAME
