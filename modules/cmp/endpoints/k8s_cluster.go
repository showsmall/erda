// Copyright (c) 2021 Terminus, Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package endpoints

import (
	"context"
	"fmt"
	"net/http"
	"strconv"

	"github.com/pkg/errors"

	"github.com/erda-project/erda/apistructs"
	"github.com/erda-project/erda/pkg/http/httpserver"
)

// ListK8SClusters list ready and unready k8s clusters in current org
func (e *Endpoints) ListK8SClusters(ctx context.Context, r *http.Request, vars map[string]string) (httpserver.Responser, error) {
	orgid := r.Header.Get("Org-ID")
	scopeID, err := strconv.ParseUint(orgid, 10, 64)
	if err != nil {
		errstr := fmt.Sprintf("failed to get org-id in http header")
		return mkResponse(apistructs.K8SClusters{
			Header: apistructs.Header{
				Success: false,
				Error: apistructs.ErrorResponse{
					Msg: errstr,
				},
			},
		})
	}
	clusters, err := e.listClusters(scopeID, "k8s", "edas")
	if err != nil {
		errstr := fmt.Sprintf("failed to list cluster, %v", err)
		return mkResponse(apistructs.K8SClusters{
			Header: apistructs.Header{
				Success: false,
				Error: apistructs.ErrorResponse{
					Msg: errstr,
				},
			},
		})
	}

	clustersInOrg := map[string]struct{}{}
	for _, cluster := range clusters {
		clustersInOrg[cluster.Name] = struct{}{}
	}
	ready, unready := e.SteveAggregator.ListClusters()

	var readyInOrg, unreadyInOrg []string
	for _, cluster := range ready {
		if _, ok := clustersInOrg[cluster]; ok {
			readyInOrg = append(readyInOrg, cluster)
		}
	}
	for _, cluster := range unready {
		if _, ok := clustersInOrg[cluster]; ok {
			unreadyInOrg = append(unreadyInOrg, cluster)
		}
	}
	return mkResponse(apistructs.K8SClusters{
		Header: apistructs.Header{
			Success: true,
		},
		Data: apistructs.ClustersData{
			Ready:   readyInOrg,
			UnReady: unreadyInOrg,
		},
	})
}

func (e *Endpoints) listClusters(scopeID uint64, clusterTypes ...string) ([]apistructs.ClusterInfo, error) {
	var clusters []apistructs.ClusterInfo
	for _, typ := range clusterTypes {
		c, err := e.bdl.ListClusters(typ, scopeID)
		if err != nil {
			return nil, errors.Errorf("failed to list %s clusters, %v", typ, err)
		}
		clusters = append(clusters, c...)
	}
	return clusters, nil
}
