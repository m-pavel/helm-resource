package cmd

import (
	"io"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	appsv1 "k8s.io/api/apps/v1"
	cv1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/yaml"
)

func TestParse_Deployment(t *testing.T) {
	f, err := os.OpenFile("../testdata/deployment.yaml", os.O_RDONLY, 0644)
	require.NoError(t, err)
	assert.NotNil(t, f)
	defer f.Close()
	date, err := io.ReadAll(f)
	require.NoError(t, err)
	s := sumCmd{}
	cr, err := s.Parse(date)
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
	s := sumCmd{}
	cr, err := s.Parse(date)
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
	s := sumCmd{}
	cr, err := s.Parse(date)
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
	s := sumCmd{}
	cr, err := s.Parse(date)
	require.NoError(t, err)
	assert.True(t, cr.Limits[jobCpu].Equal(resource.MustParse("400m")), cr.Limits.Cpu())
	assert.True(t, cr.Limits[jobMemory].Equal(resource.MustParse("2048Mi")), cr.Limits.Memory())
	assert.True(t, cr.Requests[jobCpu].Equal(resource.MustParse("250m")), cr.Requests.Cpu())
	assert.True(t, cr.Requests[jobMemory].Equal(resource.MustParse("900Mi")), cr.Requests.Memory())
}

func TestParseDefaults_Deployment(t *testing.T) {
	depl := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind: "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: appsv1.DeploymentSpec{
			Template: cv1.PodTemplateSpec{
				Spec: cv1.PodSpec{
					Containers: []cv1.Container{
						{
							Name: "test-container",
							Resources: cv1.ResourceRequirements{
								Limits: cv1.ResourceList{
									cv1.ResourceCPU:    resource.MustParse("500m"),
									cv1.ResourceMemory: resource.MustParse("1600Mi"),
								},
							},
						},
					},
				},
			},
		},
	}

	dbytes, err := yaml.Marshal(depl)
	require.NoError(t, err)
	s := sumCmd{
		baseHelmCmd: baseHelmCmd{
			require: true,
		},
	}
	_, err = s.Parse(dbytes)
	require.Error(t, err)
}

func TestParseDefaults2_Deployment(t *testing.T) {
	depl := appsv1.Deployment{
		TypeMeta: metav1.TypeMeta{
			Kind: "Deployment",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: "test",
		},
		Spec: appsv1.DeploymentSpec{
			Template: cv1.PodTemplateSpec{
				Spec: cv1.PodSpec{
					Containers: []cv1.Container{
						{
							Name:      "test-container",
							Resources: cv1.ResourceRequirements{},
						},
					},
				},
			},
		},
	}

	dbytes, err := yaml.Marshal(depl)
	require.NoError(t, err)
	s := sumCmd{
		baseHelmCmd: baseHelmCmd{
			require:         true,
			defaultCpuLimit: "1",
			defaultMemLimit: "1Gi",
			defaultCpuReq:   "2",
			defaultMemReq:   "2Gi",
		},
	}
	cr, err := s.Parse(dbytes)
	require.NoError(t, err)
	assert.True(t, cr.Limits.Cpu().Equal(resource.MustParse("1")), cr.Limits.Cpu())
	assert.True(t, cr.Limits.Memory().Equal(resource.MustParse("1Gi")), cr.Limits.Memory())
	assert.True(t, cr.Requests.Cpu().Equal(resource.MustParse("2")), cr.Requests.Cpu())
	assert.True(t, cr.Requests.Memory().Equal(resource.MustParse("2Gi")), cr.Requests.Memory())
}
