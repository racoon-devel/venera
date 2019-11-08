package classifier

import (
	"bufio"
	tf "github.com/tensorflow/tensorflow/tensorflow/go"
	"io/ioutil"
	"os"
)

func loadGraph(graphFile string) (*tf.Graph, error) {
	var err error
	graphModel, err := ioutil.ReadFile(graphFile)
	if err != nil {
		return nil, err
	}

	graph := tf.NewGraph()
	if err := graph.Import(graphModel, ""); err != nil {
		return nil, err
	}

	return graph, nil
}

func loadLabels(labelsFile string) ([]string, error) {
	file, err := os.Open(labelsFile)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	labels := make([]string, 0)
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		labels = append(labels, scanner.Text())
	}

	if err := scanner.Err(); err != nil {
		return nil, err
	}

	return labels, err
}
