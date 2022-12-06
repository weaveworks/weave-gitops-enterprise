package server_test

import (
	"context"
	"log"
	"net"
	"testing"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	pb "github.com/weaveworks/weave-gitops-enterprise/pkg/api/applications"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/applications/server"
	"github.com/weaveworks/weave-gitops/pkg/git/gitfakes"
	"github.com/weaveworks/weave-gitops/pkg/gitproviders/gitprovidersfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/auth"
	"github.com/weaveworks/weave-gitops/pkg/services/auth/authfakes"
	"github.com/weaveworks/weave-gitops/pkg/services/servicesfakes"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/test/bufconn"
	"k8s.io/apimachinery/pkg/util/rand"
)

func TestServer(t *testing.T) {
	RegisterFailHandler(Fail)

	RunSpecs(t, "Server")
}

const bufSize = 1024 * 1024

var lis *bufconn.Listener

var s *grpc.Server
var apps pb.ApplicationsServer
var appsClient pb.ApplicationsClient
var conn *grpc.ClientConn
var ghAuthClient *authfakes.FakeGithubAuthClient
var gitProvider *gitprovidersfakes.FakeGitProvider
var glAuthClient *authfakes.FakeGitlabAuthClient
var configGit *gitfakes.FakeGit
var fakeFactory *servicesfakes.FakeFactory
var jwtClient auth.JWTClient

func bufDialer(context.Context, string) (net.Conn, error) {
	return lis.Dial()
}

var secretKey string

var _ = BeforeEach(func() {
	lis = bufconn.Listen(bufSize)
	s = grpc.NewServer()

	rand.Seed(time.Now().UnixNano())
	secretKey = rand.String(20)

	gitProvider = &gitprovidersfakes.FakeGitProvider{}
	gitProvider.GetDefaultBranchReturns("main", nil)

	fakeFactory = &servicesfakes.FakeFactory{}
	configGit = &gitfakes.FakeGit{}

	fakeFactory.GetGitClientsReturns(configGit, gitProvider, nil)

	ghAuthClient = &authfakes.FakeGithubAuthClient{}
	glAuthClient = &authfakes.FakeGitlabAuthClient{}
	jwtClient = auth.NewJwtClient(secretKey)

	cfg := server.ApplicationsConfig{
		JwtClient:        jwtClient,
		GithubAuthClient: ghAuthClient,
		GitlabAuthClient: glAuthClient,
	}
	apps = server.NewApplicationsServer(&cfg)
	pb.RegisterApplicationsServer(s, apps)

	go func() {
		if err := s.Serve(lis); err != nil {
			log.Fatalf(err.Error())
		}
	}()

	ctx := context.Background()
	var err error
	conn, err = grpc.DialContext(ctx, "bufnet", grpc.WithContextDialer(bufDialer), grpc.WithTransportCredentials(insecure.NewCredentials()))
	Expect(err).NotTo(HaveOccurred())

	appsClient = pb.NewApplicationsClient(conn)
})

var _ = AfterEach(func() {
	conn.Close()
})
