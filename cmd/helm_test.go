package cmd

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestParse_Deployment(t *testing.T) {
	f, err := os.OpenFile("../testdata/deployment.yaml", os.O_RDONLY, 0644)
	require.NoError(t, err)
	assert.NotNil(t, f)
	defer f.Close()
	date, err := io.ReadAll(f)
	require.NoError(t, err)
	cr, err := Parse(date)
	require.NoError(t, err)
	assert.True(t, cr.Limits.Cpu().Equal(resource.MustParse("500m")), cr.Limits.Cpu())
	assert.True(t, cr.Limits.Memory().Equal(resource.MustParse("1600Mi")), cr.Limits.Memory())
	assert.True(t, cr.Requests.Cpu().Equal(resource.MustParse("250m")), cr.Requests.Cpu())
	assert.True(t, cr.Requests.Memory().Equal(resource.MustParse("1510Mi")), cr.Requests.Memory())
}

func TestParse_Deployment_Repl(t *testing.T) {
	f, err := os.OpenFile("../testdata/deployment2.yaml", os.O_RDONLY, 0644)
	require.NoError(t, err)
	assert.NotNil(t, f)
	defer f.Close()
	date, err := io.ReadAll(f)
	require.NoError(t, err)
	cr, err := Parse(date)
	require.NoError(t, err)
	assert.True(t, cr.Limits.Cpu().Equal(resource.MustParse("1350m")), cr.Limits.Cpu())
	assert.True(t, cr.Limits.Memory().Equal(resource.MustParse("4500Mi")), cr.Limits.Memory())
	assert.True(t, cr.Requests.Cpu().Equal(resource.MustParse("600m")), cr.Requests.Cpu())
	assert.True(t, cr.Requests.Memory().Equal(resource.MustParse("4500Mi")), cr.Requests.Memory())
}

func TestParse_STatefullSet(t *testing.T) {
	f, err := os.OpenFile("../testdata/ss.yaml", os.O_RDONLY, 0644)
	require.NoError(t, err)
	assert.NotNil(t, f)
	defer f.Close()
	date, err := io.ReadAll(f)
	require.NoError(t, err)
	cr, err := Parse(date)
	require.NoError(t, err)
	assert.True(t, cr.Limits.Cpu().Equal(resource.MustParse("500m")), cr.Limits.Cpu())
	assert.True(t, cr.Limits.Memory().Equal(resource.MustParse("2000Mi")), cr.Limits.Memory())
	assert.True(t, cr.Requests.Cpu().Equal(resource.MustParse("200m")), cr.Requests.Cpu())
	assert.True(t, cr.Requests.Memory().Equal(resource.MustParse("1000Mi")), cr.Requests.Memory())
}

func TestParse_Job(t *testing.T) {
	f, err := os.OpenFile("../testdata/cj.yaml", os.O_RDONLY, 0644)
	require.NoError(t, err)
	assert.NotNil(t, f)
	defer f.Close()
	date, err := io.ReadAll(f)
	require.NoError(t, err)
	cr, err := Parse(date)
	require.NoError(t, err)
	assert.True(t, cr.Limits.Cpu().Equal(resource.MustParse("400m")), cr.Limits.Cpu())
	assert.True(t, cr.Limits.Memory().Equal(resource.MustParse("2048Mi")), cr.Limits.Memory())
	assert.True(t, cr.Requests.Cpu().Equal(resource.MustParse("250m")), cr.Requests.Cpu())
	assert.True(t, cr.Requests.Memory().Equal(resource.MustParse("900Mi")), cr.Requests.Memory())
}
