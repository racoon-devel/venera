package classifier

import (
	"bytes"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
	"golang.org/x/image/bmp"
	"image"
)

func createTensorFromImage(imageBuffer *bytes.Buffer) (*tf.Tensor, error) {
	tensor, err := tf.NewTensor(imageBuffer.String())
	if err != nil {
		return nil, err
	}
	graph, input, output, err := createTransformImageGraph()
	if err != nil {
		return nil, err
	}
	session, err := tf.NewSession(graph, nil)
	if err != nil {
		return nil, err
	}
	defer session.Close()
	normalized, err := session.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, err
	}
	return normalized[0], nil
}



func createTransformImageGraph() (graph *tf.Graph, input, output tf.Output, err error) {
	const (
		H, W  = 299, 299
		Mean  = float32(0)
		Scale = float32(255)
	)
	s := op.NewScope()
	input = op.Placeholder(s, tf.String)
	decode := op.DecodeBmp(s, input, op.DecodeBmpChannels(3))

	output = op.Div(s,
		op.Sub(s,
			op.ResizeBilinear(s,
				op.ExpandDims(s,
						op.Cast(s, decode, tf.Float),
					op.Const(s.SubScope("make_batch"), int32(0))),
				op.Const(s.SubScope("size"), []int32{H, W})),
			op.Const(s.SubScope("mean"), Mean)),
		op.Const(s.SubScope("scale"), Scale))
	graph, err = s.Finalize()
	return graph, input, output, err
}


func (c *Classifier) Classify(preparedImage image.Image) (float32, error) {
	var buf bytes.Buffer
	if err := bmp.Encode(&buf, preparedImage); err != nil {
		return 0, err
	}

	tensor, err := createTensorFromImage(&buf)
	if err != nil {
		return 0, err
	}

	output, err := c.main.session.Run(
		map[tf.Output]*tf.Tensor{
			c.input.Output(0): tensor,
		},
		[]tf.Output{
			c.output.Output(0),
		},
		nil)

	if err != nil {
		return 0, err
	}

	rating := output[0].Value().([][]float32)[0][1]
	return rating, nil
}
