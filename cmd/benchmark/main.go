package main

import (
	"flag"
	"fmt"
	"github.com/prometheus/common/promslog"
	"github.com/prometheus/pushgateway/storage"
	"github.com/prometheus/pushgateway/testutil"
	"os"
	"runtime/pprof"

	"github.com/golang/protobuf/proto"
	dto "github.com/prometheus/client_model/go"
	"time"
)

var (
	logger = promslog.NewNopLogger()
	mf3    = &dto.MetricFamily{
		Name: proto.String("mf3"),
		Type: dto.MetricType_UNTYPED.Enum(),
		Metric: []*dto.Metric{
			{
				Label: []*dto.LabelPair{
					{
						Name:  proto.String("instance"),
						Value: proto.String("instance1"),
					},
					{
						Name:  proto.String("job"),
						Value: proto.String("job1"),
					},
				},
				Untyped: &dto.Untyped{
					Value: proto.Float64(42),
				},
			},
		},
	}
)

func oneSubmit(dms *storage.DiskMetricStore, i int) {
	ts1 := time.Now()
	grouping1 := map[string]string{
		"job":      "job1",
		"instance": "instance1",
	}
	errCh := make(chan error, 1)
	name := fmt.Sprintf("Mf%d", i)
	mf3.Name = &name

	dms.SubmitWriteRequest(storage.WriteRequest{
		Labels:         grouping1,
		Timestamp:      ts1,
		MetricFamilies: testutil.MetricFamiliesMap(mf3),
		Done:           errCh,
	})
	for err := range errCh {
		panic(err)
	}

}

var cpuprofile = flag.String("cpuprofile", "", "write cpu profile to file")

func main() {
	flag.Parse()
	if *cpuprofile != "" {
		f, err := os.Create(*cpuprofile)
		if err != nil {
			panic(err)
		}
		pprof.StartCPUProfile(f)
		defer pprof.StopCPUProfile()
	}

	t1 := time.Now()
	last := t1
	dms := storage.NewDiskMetricStore("", 100*time.Millisecond, nil, logger)
	for i := 1; i < 3000; i++ {
		oneSubmit(dms, i)
		if i%100 == 0 {
			t := time.Now()
			elapsed := t.Sub(t1)
			diff := t.Sub(last)
			last = t
			fmt.Printf("I: %d elapsed:%s diff:%s\n", i, elapsed, diff)
		}
	}

	fmt.Println("Big success")
}
