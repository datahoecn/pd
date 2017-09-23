// Copyright 2016 PingCAP, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//	   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"fmt"

	. "github.com/pingcap/check"
	"github.com/pingcap/kvproto/pkg/metapb"
	"github.com/pingcap/pd/server/core"
)

var _ = Suite(&testKVSuite{})

type testKVSuite struct {
	server  *Server
	cleanup cleanUpFunc
}

func (s *testKVSuite) SetUpTest(c *C) {
	s.server, s.cleanup = mustRunTestServer(c)
}

func (s *testKVSuite) TearDownTest(c *C) {
	s.cleanup()
}

func (s *testKVSuite) TestBasic(c *C) {
	kv := newKV(s.server)

	clusterID := s.server.clusterID
	storePath := fmt.Sprintf("/pd/%v/raft/s/00000000000000000123", clusterID)
	regionPath := fmt.Sprintf("/pd/%v/raft/r/00000000000000000123", clusterID)
	c.Assert(kv.storePath(123), Equals, storePath)
	c.Assert(kv.regionPath(123), Equals, regionPath)

	meta := &metapb.Cluster{Id: 123}
	ok, err := kv.loadMeta(meta)
	c.Assert(ok, IsFalse)
	c.Assert(err, IsNil)
	c.Assert(kv.saveMeta(meta), IsNil)
	newMeta := &metapb.Cluster{}
	ok, err = kv.loadMeta(newMeta)
	c.Assert(ok, IsTrue)
	c.Assert(err, IsNil)
	c.Assert(newMeta, DeepEquals, meta)

	store := &metapb.Store{Id: 123}
	ok, err = kv.loadStore(123, store)
	c.Assert(ok, IsFalse)
	c.Assert(err, IsNil)
	c.Assert(kv.saveStore(store), IsNil)
	newStore := &metapb.Store{}
	ok, err = kv.loadStore(123, newStore)
	c.Assert(ok, IsTrue)
	c.Assert(err, IsNil)
	c.Assert(newStore, DeepEquals, store)

	region := &metapb.Region{Id: 123}
	ok, err = kv.loadRegion(123, region)
	c.Assert(ok, IsFalse)
	c.Assert(err, IsNil)
	c.Assert(kv.saveRegion(region), IsNil)
	newRegion := &metapb.Region{}
	ok, err = kv.loadRegion(123, newRegion)
	c.Assert(ok, IsTrue)
	c.Assert(err, IsNil)
	c.Assert(newRegion, DeepEquals, region)
}

func mustSaveStores(c *C, kv *kv, n int) []*metapb.Store {
	stores := make([]*metapb.Store, 0, n)
	for i := 0; i < n; i++ {
		store := &metapb.Store{Id: uint64(i)}
		stores = append(stores, store)
	}

	for _, store := range stores {
		c.Assert(kv.saveStore(store), IsNil)
	}

	return stores
}

func (s *testKVSuite) TestLoadStores(c *C) {
	kv := newKV(s.server)
	cache := core.NewStoresInfo()

	n := 10
	stores := mustSaveStores(c, kv, n)
	c.Assert(kv.loadStores(cache, 3), IsNil)

	c.Assert(cache.GetStoreCount(), Equals, n)
	for _, store := range cache.GetMetaStores() {
		c.Assert(store, DeepEquals, stores[store.GetId()])
	}
}

func (s *testKVSuite) TestStoreWeight(c *C) {
	kv := newKV(s.server)
	cache := core.NewStoresInfo()
	const n = 3

	mustSaveStores(c, kv, n)
	c.Assert(kv.saveStoreWeight(1, 2.0, 3.0), IsNil)
	c.Assert(kv.saveStoreWeight(2, 0.2, 0.3), IsNil)
	c.Assert(kv.loadStores(cache, n), IsNil)
	leaderWeights := []float64{1.0, 2.0, 0.2}
	regionWeights := []float64{1.0, 3.0, 0.3}
	for i := 0; i < n; i++ {
		c.Assert(cache.GetStore(uint64(i)).LeaderWeight, Equals, leaderWeights[i])
		c.Assert(cache.GetStore(uint64(i)).RegionWeight, Equals, regionWeights[i])
	}
}

func mustSaveRegions(c *C, kv *kv, n int) []*metapb.Region {
	regions := make([]*metapb.Region, 0, n)
	for i := 0; i < n; i++ {
		region := &metapb.Region{Id: uint64(i)}
		regions = append(regions, region)
	}

	for _, region := range regions {
		c.Assert(kv.saveRegion(region), IsNil)
	}

	return regions
}

func (s *testKVSuite) TestLoadRegions(c *C) {
	kv := newKV(s.server)
	cache := core.NewRegionsInfo()

	n := 10
	regions := mustSaveRegions(c, kv, n)
	c.Assert(kv.loadRegions(cache, 3), IsNil)

	c.Assert(cache.GetRegionCount(), Equals, n)
	for _, region := range cache.GetMetaRegions() {
		c.Assert(region, DeepEquals, regions[region.GetId()])
	}
}
