package database_test

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/weaveworks/weave-gitops-enterprise/pkg/handlers/api/database"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/models"
	"github.com/weaveworks/weave-gitops-enterprise/common/database/utils"
)

func TestGetClusters_NilDb(t *testing.T) {
	_, err := database.GetClusters(nil, database.GetClustersRequest{})
	assert.Equal(t, database.ErrNilDB, err)
}

func TestGetClusters_Pagination(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)
	total := 101
	for i := 1; i <= total; i++ {
		db.Create(&models.Cluster{
			Name:  fmt.Sprintf("Cluster %03d", i),
			Token: fmt.Sprintf("Token %03d", i),
		})
	}

	records001to010 := []string{
		"Cluster 001",
		"Cluster 002",
		"Cluster 003",
		"Cluster 004",
		"Cluster 005",
		"Cluster 006",
		"Cluster 007",
		"Cluster 008",
		"Cluster 009",
		"Cluster 010",
	}

	records011to020 := []string{
		"Cluster 011",
		"Cluster 012",
		"Cluster 013",
		"Cluster 014",
		"Cluster 015",
		"Cluster 016",
		"Cluster 017",
		"Cluster 018",
		"Cluster 019",
		"Cluster 020",
	}

	records001to050 := []string{
		"Cluster 001",
		"Cluster 002",
		"Cluster 003",
		"Cluster 004",
		"Cluster 005",
		"Cluster 006",
		"Cluster 007",
		"Cluster 008",
		"Cluster 009",
		"Cluster 010",
		"Cluster 011",
		"Cluster 012",
		"Cluster 013",
		"Cluster 014",
		"Cluster 015",
		"Cluster 016",
		"Cluster 017",
		"Cluster 018",
		"Cluster 019",
		"Cluster 020",
		"Cluster 021",
		"Cluster 022",
		"Cluster 023",
		"Cluster 024",
		"Cluster 025",
		"Cluster 026",
		"Cluster 027",
		"Cluster 028",
		"Cluster 029",
		"Cluster 030",
		"Cluster 031",
		"Cluster 032",
		"Cluster 033",
		"Cluster 034",
		"Cluster 035",
		"Cluster 036",
		"Cluster 037",
		"Cluster 038",
		"Cluster 039",
		"Cluster 040",
		"Cluster 041",
		"Cluster 042",
		"Cluster 043",
		"Cluster 044",
		"Cluster 045",
		"Cluster 046",
		"Cluster 047",
		"Cluster 048",
		"Cluster 049",
		"Cluster 050",
	}

	records051to100 := []string{
		"Cluster 051",
		"Cluster 052",
		"Cluster 053",
		"Cluster 054",
		"Cluster 055",
		"Cluster 056",
		"Cluster 057",
		"Cluster 058",
		"Cluster 059",
		"Cluster 060",
		"Cluster 061",
		"Cluster 062",
		"Cluster 063",
		"Cluster 064",
		"Cluster 065",
		"Cluster 066",
		"Cluster 067",
		"Cluster 068",
		"Cluster 069",
		"Cluster 070",
		"Cluster 071",
		"Cluster 072",
		"Cluster 073",
		"Cluster 074",
		"Cluster 075",
		"Cluster 076",
		"Cluster 077",
		"Cluster 078",
		"Cluster 079",
		"Cluster 080",
		"Cluster 081",
		"Cluster 082",
		"Cluster 083",
		"Cluster 084",
		"Cluster 085",
		"Cluster 086",
		"Cluster 087",
		"Cluster 088",
		"Cluster 089",
		"Cluster 090",
		"Cluster 091",
		"Cluster 092",
		"Cluster 093",
		"Cluster 094",
		"Cluster 095",
		"Cluster 096",
		"Cluster 097",
		"Cluster 098",
		"Cluster 099",
		"Cluster 100",
	}

	var records001to100 []string
	records001to100 = append(records001to100, records001to050...)
	records001to100 = append(records001to100, records051to100...)

	// Size equivalence classes: 9,10,50,100,101
	// Page equivalence classes: 0,1,2
	tests := []struct {
		name    string
		req     database.GetClustersRequest
		records []string
		err     error
	}{
		{
			name: "size < 10 && page < 1 returns 10 records (001-010)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 9,
					Page: 0,
				},
			},
			records: records001to010,
		},
		{
			name: "size < 10 && page = 1 returns 10 records (001-010)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 9,
					Page: 1,
				},
			},
			records: records001to010,
		},
		{
			name: "size < 10 && page > 1 returns 10 records (011-020)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 9,
					Page: 2,
				},
			},
			records: records011to020,
		},
		{
			name: "size = 10 && page < 1 returns 10 records (001-010)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 10,
					Page: 0,
				},
			},
			records: records001to010,
		},
		{
			name: "size = 10 && page = 1 returns 10 records (001-010)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 10,
					Page: 1,
				},
			},
			records: records001to010,
		},
		{
			name: "size = 10 && page > 1 returns 10 records (011-020)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 10,
					Page: 2,
				},
			},
			records: records011to020,
		},
		{
			name: "10 < size < 100 && page < 1 returns 50 records (001-050)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 50,
					Page: 0,
				},
			},
			records: records001to050,
		},
		{
			name: "10 < size < 100 && page = 1 returns 50 records (001-050)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 50,
					Page: 1,
				},
			},
			records: records001to050,
		},
		{
			name: "10 < size < 100 && page > 1 returns 50 records (051-100)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 50,
					Page: 2,
				},
			},
			records: records051to100,
		},
		{
			name: "size = 100 && page < 1 returns 100 records (001-100)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 100,
					Page: 0,
				},
			},
			records: records001to100,
		},
		{
			name: "size = 100 && page = 0 returns 100 records (001-100)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 100,
					Page: 1,
				},
			},
			records: records001to100,
		},
		{
			name: "size = 100 && page > 1 returns 1 record (101)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 100,
					Page: 2,
				},
			},
			records: []string{"Cluster 101"},
		},
		{
			name: "size > 100 && page < 1 returns 100 records (001-100)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 101,
					Page: 0,
				},
			},
			records: records001to100,
		},
		{
			name: "size > 100 && page = 0 returns 100 records (001-100)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 101,
					Page: 1,
				},
			},
			records: records001to100,
		},
		{
			name: "size > 100 && page > 1 returns 1 record (101)",
			req: database.GetClustersRequest{
				Pagination: database.Pagination{
					Size: 101,
					Page: 2,
				},
			},
			records: []string{"Cluster 101"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			res, err := database.GetClusters(db, tt.req)
			assert.Equal(t, tt.err, err)
			assert.Equal(t, len(tt.records), len(res.Clusters))
			var names []string
			for _, c := range res.Clusters {
				names = append(names, c.Name)
			}
			assert.EqualValues(t, tt.records, names)
			assert.Equal(t, int64(total), res.Total)
		})
	}
}

// https://weaveworks.atlassian.net/browse/WKP-2224
func TestGetClusters_CAPIClusterWithMultipleVersions(t *testing.T) {
	db, err := utils.Open("", "sqlite", "", "", "")
	require.NoError(t, err)
	err = utils.MigrateTables(db)
	require.NoError(t, err)

	// Create a cluster with > 1 different CAPI versions
	db.Create(&models.Cluster{
		Name:          "test-cluster",
		Token:         "test-token",
		CAPIName:      "capi-name",
		CAPINamespace: "capi-namespace",
	})
	for i := 1; i <= 2; i++ {
		capiCluster := &models.CAPICluster{
			Name:        "capi-name",
			Namespace:   "capi-namespace",
			CAPIVersion: fmt.Sprintf("%d", i),
		}
		db.Create(capiCluster)
	}

	// Create at least 10 more clusters (default pagination size is 10)
	for i := 1; i <= 10; i++ {
		db.Create(&models.Cluster{
			Name:  fmt.Sprintf("test-cluster-%d", i),
			Token: fmt.Sprintf("test-token-%d", i),
		})
	}

	res, err := database.GetClusters(db, database.GetClustersRequest{})
	assert.NoError(t, err)
	assert.Equal(t, 10, len(res.Clusters))
	assert.Equal(t, int64(11), res.Total)
}
