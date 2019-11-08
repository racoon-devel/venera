package classifier

import (
	"bytes"
	"fmt"
	"github.com/oliamb/cutter"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"github.com/tensorflow/tensorflow/tensorflow/go/op"
	"image"
	"image/color"
	_ "image/jpeg"
)

const (
	personClass float32 = 1
	personThreshold float32 = 0.5
)

func (c *Classifier) PrepareImage(encodedImage []byte) (image.Image, error) {
	tensor, img, err := makeTensorFromImage(encodedImage, 3)
	if err != nil {
		return nil, fmt.Errorf("error making input tensor: %v", err)
	}

	probabilities, classes, boxes, err := detectObjects(c.detector.session, c.detector.graph, tensor)
	if err != nil {
		return nil, fmt.Errorf("error making prediction: %v", err)
	}

	persons := make([][]float32, 0)

	for i, box := range boxes {
		score := probabilities[i]
		class := classes[i]
		if score > personThreshold && class == personClass {
			persons = append(persons, box)
		}
	}

	if len(persons) != 0 {
		person := persons[0]
		bounds := img.Bounds()
		y1 := int(float64(bounds.Min.Y) + float64(bounds.Dy())*float64(person[0]))
		x1 := int(float64(bounds.Min.X) + float64(bounds.Dx())*float64(person[1]))
		y2 := int(float64(bounds.Min.Y) + float64(bounds.Dy())*float64(person[2]))
		x2 := int(float64(bounds.Min.X) + float64(bounds.Dx())*float64(person[3]))

		cropped, err := cutter.Crop(img, cutter.Config{
			Width:   x2-x1,
			Height:  y2-y1,
			Anchor:  image.Point{x1, y1},
			Mode:    cutter.TopLeft,
			Options: 0,
		})

		if err != nil {
			return nil, err
		}

		return convertToGreyscale(cropped), nil
	}

	return nil, nil
}

func makeTensorFromImage(img []byte, channels int) (*tf.Tensor, image.Image, error) {
	tensor, err := tf.NewTensor(string(img))
	if err != nil {
		return nil, nil, err
	}
	normalizeGraph, input, output, err := decodeBitmapGraph(channels)
	if err != nil {
		return nil, nil, err
	}
	normalizeSession, err := tf.NewSession(normalizeGraph, nil)
	if err != nil {
		return nil, nil, err
	}
	defer normalizeSession.Close()
	normalized, err := normalizeSession.Run(
		map[tf.Output]*tf.Tensor{input: tensor},
		[]tf.Output{output},
		nil)
	if err != nil {
		return nil, nil, err
	}

	r := bytes.NewReader(img)
	i, _, err := image.Decode(r)
	if err != nil {
		return nil, nil, err
	}
	return normalized[0], i, nil
}

func decodeBitmapGraph(channels int) (*tf.Graph, tf.Output, tf.Output, error) {
	s := op.NewScope()
	input := op.Placeholder(s, tf.String)
	output := op.ExpandDims(
		s,
		op.DecodeJpeg(s, input, op.DecodeJpegChannels(int64(channels))),
		op.Const(s.SubScope("make_batch"), int32(0)))
	graph, err := s.Finalize()
	return graph, input, output, err
}

func detectObjects(session *tf.Session, graph *tf.Graph, input *tf.Tensor) ([]float32, []float32, [][]float32, error) {
	inputop := graph.Operation("image_tensor")
	output, err := session.Run(
		map[tf.Output]*tf.Tensor{
			inputop.Output(0): input,
		},
		[]tf.Output{
			graph.Operation("detection_boxes").Output(0),
			graph.Operation("detection_scores").Output(0),
			graph.Operation("detection_classes").Output(0),
			graph.Operation("num_detections").Output(0),
		},
		nil)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("Error running session: %v", err)
	}
	probabilities := output[1].Value().([][]float32)[0]
	classes := output[2].Value().([][]float32)[0]
	boxes := output[0].Value().([][][]float32)[0]
	return probabilities, classes, boxes, nil
}

func convertToGreyscale(img image.Image) image.Image {
	b := img.Bounds()
	imgSet := image.NewRGBA(b)
	for y := 0; y < b.Max.Y; y++ {
		for x := 0; x < b.Max.X; x++ {
			oldPixel := img.At(x, y)
			imgSet.Set(x, int(y), color.GrayModel.Convert(oldPixel))
		}
	}

	return imgSet
}
