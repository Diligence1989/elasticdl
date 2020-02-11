package ps

import (
	"elasticdl.org/elasticdl/pkg/common"
	"elasticdl.org/elasticdl/pkg/proto"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestSGDOptimizer(t *testing.T) {
	d1 := []int64{2, 3}
	v1 := []float32{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}
	t1 := common.NewTensor(v1, d1) //t1

	d2 := []int64{2, 2}
	v2 := []float32{1.0, 2.0, 1.1, 2.2}
	t2 := common.NewTensor(v2, d2) //t2

	model := NewModel()
	model.DenseParameters["t1"] = t1
	model.DenseParameters["t2"] = t2

	gv1 := []float32{1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
	gv2 := []float32{1.0, 1.0, 1.0, 1.0}
	grad1 := common.NewTensor(gv1, d1) //t1
	grad2 := common.NewTensor(gv2, d2) //t2
	pbModel := &proto.Model{
		DenseParameters: map[string]*proto.Tensor{"t1": grad1, "t2": grad2},
	}

	opt := NewSGDOptimizer(0.1)

	// test dense parameter update
	err1 := opt.ApplyGradients(pbModel, model)
	assert.Equal(t, opt.GetLR(), float32(0.1))
	assert.Nil(t, err1)

	ev1 := []float32{0.9, 1.9, 2.9, 3.9, 4.9, 5.9}
	ev2 := []float32{0.9, 1.9, 1.0, 2.1}

	assert.True(t, common.CompareFloatArray(common.Slice(model.DenseParameters["t1"]).([]float32), ev1, 0.0001))
	assert.True(t, common.CompareFloatArray(common.Slice(model.DenseParameters["t2"]).([]float32), ev2, 0.0001))

	// test grad name error
	grad3 := common.NewTensor(gv2, d2) //t3
	pbModel = &proto.Model{
		DenseParameters: map[string]*proto.Tensor{"t3": grad3},
	}
	err2 := opt.ApplyGradients(pbModel, model)
	assert.NotNil(t, err2)

	// test sparse parameter update
	info := &proto.EmbeddingTableInfo{
		Name:        "t3",
		Dim:         2,
		Initializer: "zero",
		Dtype:       common.Float32,
	}
	model.SetEmbeddingTableInfo(info)

	d3 := []int64{2, 2}
	v3 := []float32{1.0, 1.0, 1.0, 1.0}
	i3 := []int64{1, 3}
	grad3 = common.NewTensor(v3, d3) // t3
	sgrad3 := common.NewIndexedSlices(grad3, i3)

	pbModel = &proto.Model{
		DenseParameters: map[string]*proto.Tensor{"t1": grad1, "t2": grad2},
		IndexedSlices:   map[string]*proto.IndexedSlices{"t3": sgrad3},
	}

	err3 := opt.ApplyGradients(pbModel, model)
	assert.Nil(t, err3)

	ev1 = []float32{0.8, 1.8, 2.8, 3.8, 4.8, 5.8}
	ev2 = []float32{0.8, 1.8, 0.9, 2.0}
	assert.True(t, common.CompareFloatArray(common.Slice(model.DenseParameters["t1"]).([]float32), ev1, 0.0001))
	assert.True(t, common.CompareFloatArray(common.Slice(model.DenseParameters["t2"]).([]float32), ev2, 0.0001))

	vectors := model.GetEmbeddingTable("t3").GetEmbeddingVectors(i3)
	expV := []float32{-0.1, -0.1, -0.1, -0.1}
	assert.True(t, common.CompareFloatArray(expV, common.Slice(vectors).([]float32), 0.0001))

	// more test for sparse parameter update
	d3 = []int64{4, 2}
	v3 = []float32{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
	i3 = []int64{1, 3, 3, 5}
	grad3 = common.NewTensor(v3, d3) // t3
	sgrad3 = common.NewIndexedSlices(grad3, i3)

	pbModel = &proto.Model{
		IndexedSlices: map[string]*proto.IndexedSlices{"t3": sgrad3},
	}

	err4 := opt.ApplyGradients(pbModel, model)
	assert.Nil(t, err4)

	vectors = model.GetEmbeddingTable("t3").GetEmbeddingVectors([]int64{1, 3, 5})
	expV = []float32{-0.2, -0.2, -0.3, -0.3, -0.1, -0.1}
	assert.True(t, common.CompareFloatArray(expV, common.Slice(vectors).([]float32), 0.0001))
}

func TestAdamOptimizer(t *testing.T) {
	d1 := []int64{2, 3}
	v1 := []float32{1.0, 2.0, 3.0, 4.0, 5.0, 6.0}
	t1 := common.NewTensor(v1, d1) //t1

	d2 := []int64{2, 2}
	v2 := []float32{1.0, 2.0, 1.1, 2.2}
	t2 := common.NewTensor(v2, d2) //t2

	model := NewModel()
	model.DenseParameters["t1"] = t1
	model.DenseParameters["t2"] = t2

	gv1 := []float32{1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
	gv2 := []float32{1.0, 1.0, 1.0, 1.0}
	grad1 := common.NewTensor(gv1, d1) //t1
	grad2 := common.NewTensor(gv2, d2) //t2
	pbModel := &proto.Model{
		DenseParameters: map[string]*proto.Tensor{"t1": grad1, "t2": grad2},
		EmbeddingTableInfos: []*proto.EmbeddingTableInfo{
			&proto.EmbeddingTableInfo{
				Name:        "t3",
				Dim:         2,
				Initializer: "zero",
				Dtype:       common.Float32,
			},
		},
	}

	opt := NewAdamOptimizer(0.1, 0.9, 0.999, 1e-8, false)
	opt.InitFromModelPB(pbModel)
	opt.step = 1

	// test dense parameter update
	err1 := opt.ApplyGradients(pbModel, model)
	assert.Equal(t, opt.BaseOptimizer.lr, float32(0.1))
	assert.Nil(t, err1)

	ev1 := []float32{0.9255863187, 1.9255863187, 2.9255863187, 3.9255863187, 4.9255863187, 5.9255863187}
	ev2 := []float32{0.9255863187, 1.9255863187, 1.0255863187, 2.1255863187}

	assert.True(t, common.CompareFloatArray(common.Slice(model.DenseParameters["t1"]).([]float32), ev1, 0.0001))
	assert.True(t, common.CompareFloatArray(common.Slice(model.DenseParameters["t2"]).([]float32), ev2, 0.0001))

	// test grad name error
	grad3 := common.NewTensor(gv2, d2) //t3
	pbModel = &proto.Model{
		DenseParameters: map[string]*proto.Tensor{"t3": grad3},
	}
	err2 := opt.ApplyGradients(pbModel, model)
	assert.NotNil(t, err2)

	// test sparse parameter update
	info := &proto.EmbeddingTableInfo{
		Name:        "t3",
		Dim:         2,
		Initializer: "zero",
		Dtype:       common.Float32,
	}
	model.SetEmbeddingTableInfo(info)

	d3 := []int64{2, 2}
	v3 := []float32{1.0, 1.0, 1.0, 1.0}
	i3 := []int64{1, 3}
	grad3 = common.NewTensor(v3, d3) // t3
	sgrad3 := common.NewIndexedSlices(grad3, i3)

	pbModel = &proto.Model{
		DenseParameters: map[string]*proto.Tensor{"t1": grad1, "t2": grad2},
		IndexedSlices:   map[string]*proto.IndexedSlices{"t3": sgrad3},
	}

	err3 := opt.ApplyGradients(pbModel, model)
	assert.Nil(t, err3)

	ev1 = []float32{0.8474920307, 1.8474920307, 2.8474920307, 3.8474920307, 4.8474920307, 5.8474920307}
	ev2 = []float32{0.8474920307, 1.8474920307, 0.9474920307, 2.0474920307}
	assert.True(t, common.CompareFloatArray(common.Slice(model.DenseParameters["t1"]).([]float32), ev1, 0.0001))
	assert.True(t, common.CompareFloatArray(common.Slice(model.DenseParameters["t2"]).([]float32), ev2, 0.0001))

	vectors := model.GetEmbeddingTable("t3").GetEmbeddingVectors(i3)
	expV := []float32{-0.058112835, -0.058112835, -0.058112835, -0.058112835}
	assert.True(t, common.CompareFloatArray(expV, common.Slice(vectors).([]float32), 0.0001))

	// more test for sparse parameter update
	d3 = []int64{4, 2}
	v3 = []float32{1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0, 1.0}
	i3 = []int64{1, 3, 3, 5}
	grad3 = common.NewTensor(v3, d3) // t3
	sgrad3 = common.NewIndexedSlices(grad3, i3)

	pbModel = &proto.Model{
		IndexedSlices: map[string]*proto.IndexedSlices{"t3": sgrad3},
	}

	err4 := opt.ApplyGradients(pbModel, model)
	assert.Nil(t, err4)

	vectors = model.GetEmbeddingTable("t3").GetEmbeddingVectors([]int64{1, 3, 5})
	expV = []float32{-0.1314178004, -0.1314178004, -0.1288654704, -0.1288654704, -0.0545489238, -0.0545489238}
	assert.True(t, common.CompareFloatArray(expV, common.Slice(vectors).([]float32), 0.0001))
}
