package data_enrichment_test

import (
	"context"
	"fmt"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/client"
	"go.temporal.io/sdk/testsuite"
	"go.temporal.io/sdk/worker"
	"sync"
	"testing"
	"time"
)

type IntegrationTestServer struct {
	key       string
	server    *testsuite.DevServer
	client    client.Client // TODO: Not neeeded as a client comes with the dev serber ..
	taskQueue string
	worker    worker.Worker

	// List of Workflows to register
	// List of Activities to register ..
}

// DataEnrichmentHappyTestSuite happy path e2e test with realistic data; can it be triggerfed parallel?
type DataEnrichmentHappyTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	its *IntegrationTestServer
}

// Below are the test case scenarios under this  suite ..
// Scenario: Normal happy with Signals ..

// Scenario: Disrupted dependency

// DataEnrichmentSadTestSuite
type DataEnrichmentSadTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	its *IntegrationTestServer
}

// Scenario: Total Prolonged failure?

func setupIntegrationDevServer(its *IntegrationTestServer) error {
	// Start the Temporal server as a background process
	server, err := testsuite.StartDevServer(context.Background(), testsuite.DevServerOptions{
		ClientOptions: &client.Options{
			//HostPort: "localhost:12345", // the go-routine version will fail ..
			Identity: its.key,
			//Namespace: "demo",
		},
	})
	fmt.Println("DevServer:", server.FrontendHostPort())
	if err != nil {
		return err
	}
	its.server = server
	//if err != nil {
	//	s.T().Fatalf("Failed to start Temporal server: %v", err)
	//	return
	//}
	//s.server = server
	//
	//// Create a client to the Temporal server
	//// Get the server address from the DevServer instance
	//serverAddr := s.server.FrontendHostPort()
	//s.T().Logf("Connecting to Temporal server at: %s", serverAddr)
	//
	//client, err := client.Dial(client.Options{
	//	HostPort: serverAddr,
	//})
	//if err != nil {
	//	s.T().Fatalf("Failed to create Temporal client: %v", err)
	//	return
	//}
	//s.client = client
	//
	//// Create a unique task queue for this test run
	//s.taskQueue = "idempotency-test-" + uuid.New().String()
	//
	//// Create a worker to process workflow and activity tasks
	//s.worker = worker.New(client, s.taskQueue, worker.Options{})
	//
	//// Register just one workflow type to avoid workflow type confusion with idempotency
	////s.worker.RegisterWorkflow(batch.FeeDeductionWorkflow)
	////s.worker.RegisterActivity(DeductFeeActivity)
	//
	//// Start the worker
	//err = s.worker.Start()
	//if err != nil {
	//	s.T().Fatalf("Failed to start worker: %v", err)
	//	return
	//}
	return nil

}

func TestParallelRun(t *testing.T) {
	var wg sync.WaitGroup
	wg.Add(2)

	go func(key string) {
		fmt.Println("RUn Scenaro:", key)
		defer wg.Done()
		s := new(DataEnrichmentHappyTestSuite)
		s.SetT(t)
		// Setup Suite ..
		s.SetupSuite()
		// Run tests ..
		s.TestHappy()
		// Clean up ..
		s.TearDownSuite()
	}("happy")

	go func(key string) {
		fmt.Println("RUn Scenaro:", key)
		defer wg.Done()
		s := new(DataEnrichmentSadTestSuite)
		s.SetT(t)
		s.SetupSuite()
		s.TestSad()
		s.TearDownSuite()
	}("sad")

	fmt.Println("Waiting ..")
	wg.Wait()
	fmt.Println("DONE!!!")
}

// Run happy + sad in parallel?
func TestDataEnrichmentWorkflow(t *testing.T) {
	//t.Parallel() // Does not seem to work?
	//t.Run("happy", func(t *testing.T) {
	suite.Run(t, new(DataEnrichmentHappyTestSuite))
	//})
	//t.Run("sad", func(t *testing.T) {
	suite.Run(t, new(DataEnrichmentSadTestSuite))
	//})
}

func (s *DataEnrichmentHappyTestSuite) SetupSuite() {
	its := &IntegrationTestServer{
		key: "happy",
	}
	err := setupIntegrationDevServer(its)
	if err != nil {
		s.T().Fatal(err)
		return
	}
	s.its = its
}

func (s *DataEnrichmentHappyTestSuite) TestHappy() {
	fmt.Println("I am Happy:", s.its.server.FrontendHostPort())
	time.Sleep(6 * time.Second)
	s.T().Error("I am Happy FAIL!!!!:", s.its.server.FrontendHostPort())
}

func (s *DataEnrichmentHappyTestSuite) TearDownSuite() {
	if s.its.server != nil {
		s.its.server.Stop()
	}
}

func (s *DataEnrichmentSadTestSuite) SetupSuite() {
	its := &IntegrationTestServer{
		key: "sad",
	}
	err := setupIntegrationDevServer(its)
	if err != nil {
		s.T().Fatal(err)
		return
	}
	s.its = its
}

func (s *DataEnrichmentSadTestSuite) TestSad() {
	fmt.Println("I am SAD:", s.its.server.FrontendHostPort())
	time.Sleep(3 * time.Second)
}

func (s *DataEnrichmentSadTestSuite) TearDownSuite() {
	if s.its.server != nil {
		s.its.server.Stop()
	}
}
