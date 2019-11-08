package classifier

import (
	"fmt"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
)

type tfModel struct {
	graph   *tf.Graph
	labels  []string
	session *tf.Session
}

type Classifier struct {
	main tfModel
	detector tfModel

	input  *tf.Operation
	output *tf.Operation
}

func NewClassifier(graphFile, labelsFile, detectorGraphFile string) (*Classifier, error) {
	c := &Classifier{}

	var err error
	if c.main.graph, err = loadGraph(graphFile); err != nil {
		return nil, err
	}

	if c.detector.graph, err = loadGraph(detectorGraphFile); err != nil {
		return nil, err
	}

	if c.main.labels, err = loadLabels(labelsFile); err != nil {
		return nil, err
	}

	c.input = c.main.graph.Operation("Placeholder")
	c.output = c.main.graph.Operation("final_result")
	if c.input == nil || c.output == nil {
		return nil, fmt.Errorf("unable to found operations in graph")
	}

	if c.main.session, err = tf.NewSession(c.main.graph, &tf.SessionOptions{}); err != nil {
		return nil, err
	}

	if c.detector.session, err = tf.NewSession(c.detector.graph, &tf.SessionOptions{}); err != nil {
		c.main.session.Close()
		return nil, err
	}

	return c, nil
}

func (c *Classifier) Close() {
	c.main.session.Close()
	c.detector.session.Close()
}
