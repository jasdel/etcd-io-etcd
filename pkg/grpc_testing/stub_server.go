// Copyright 2023 The etcd Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package grpc_testing

import (
	"context"
	"fmt"
	"net"

	"google.golang.org/grpc"
	testpb "google.golang.org/grpc/test/grpc_testing"
)

// StubServer is borrowed from the interal package of grpc-go.
// See https://github.com/grpc/grpc-go/blob/master/internal/stubserver/stubserver.go
// Since it cannot be imported directly, we have to copy and paste it here,
// and useless code for our testing is removed.

// StubServer is a server that is easy to customize within individual test
// cases.
type StubServer struct {
	// Guarantees we satisfy this interface; panics if unimplemented methods are called.
	testpb.TestServiceServer

	EmptyCallF      func(ctx context.Context, in *testpb.Empty) (*testpb.Empty, error)
	UnaryCallF      func(ctx context.Context, in *testpb.SimpleRequest) (*testpb.SimpleResponse, error)
	FullDuplexCallF func(stream testpb.TestService_FullDuplexCallServer) error

	s *grpc.Server

	// Network and Address are parameters for Listen. Defaults will be used if these are empty before Start.
	Network string
	Address string

	cleanups []func() // Lambdas executed in Stop(); populated by Start().
}

// EmptyCall is the handler for testpb.EmptyCall.
func (ss *StubServer) EmptyCall(ctx context.Context, in *testpb.Empty) (*testpb.Empty, error) {
	return ss.EmptyCallF(ctx, in)
}

// UnaryCall is the handler for testpb.UnaryCall.
func (ss *StubServer) UnaryCall(ctx context.Context, in *testpb.SimpleRequest) (*testpb.SimpleResponse, error) {
	return ss.UnaryCallF(ctx, in)
}

// FullDuplexCall is the handler for testpb.FullDuplexCall.
func (ss *StubServer) FullDuplexCall(stream testpb.TestService_FullDuplexCallServer) error {
	return ss.FullDuplexCallF(stream)
}

// Start starts the server and creates a client connected to it.
func (ss *StubServer) Start(sopts []grpc.ServerOption, dopts ...grpc.DialOption) error {
	if ss.Network == "" {
		ss.Network = "tcp"
	}
	if ss.Address == "" {
		ss.Address = "localhost:0"
	}

	lis, err := net.Listen(ss.Network, ss.Address)
	if err != nil {
		return fmt.Errorf("net.Listen(%q, %q) = %v", ss.Network, ss.Address, err)
	}
	ss.Address = lis.Addr().String()
	ss.cleanups = append(ss.cleanups, func() { lis.Close() })

	s := grpc.NewServer(sopts...)
	testpb.RegisterTestServiceServer(s, ss)
	go s.Serve(lis)
	ss.cleanups = append(ss.cleanups, s.Stop)
	ss.s = s

	return nil
}

// Stop stops ss and cleans up all resources it consumed.
func (ss *StubServer) Stop() {
	for i := len(ss.cleanups) - 1; i >= 0; i-- {
		ss.cleanups[i]()
	}
}

// Addr gets the address the server listening on.
func (ss *StubServer) Addr() string {
	return ss.Address
}