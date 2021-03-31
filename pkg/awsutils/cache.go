package awsutils

import (
	"context"
	"sync"
	"time"

	"github.com/go-logr/logr"

	// use aws sdk go v2
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

type Cache struct {
	sync.RWMutex
	logr.Logger
	ctx    context.Context
	cache  map[string]string
	awsTag string
}

func NewCache(ctx context.Context, clusterTag string) *Cache {
	return &Cache{
		ctx: ctx, cache: make(map[string]string), awsTag: clusterTag,
	}
}

func (c *Cache) InjectLogger(l logr.Logger) error {
	c.Logger = l
	return nil
}

func (c *Cache) Start() { c.start() }
func (c *Cache) start() *Cache {
	c.refresh()
	go func() {
		ticker := time.NewTicker(time.Second * 3600)
		for {
			select {
			case <-c.ctx.Done():
				ticker.Stop()
				return
			case <-ticker.C:
				c.refresh()
			}
		}
	}()
	return c
}

func (c *Cache) refresh() {
	// avoid to cache all instances in the account
	if c.awsTag == "" {
		return
	}
	ec2svc := GetEC2ServiceOrDie(c.ctx)
	p := ec2.NewDescribeInstancesPaginator(ec2svc, &ec2.DescribeInstancesInput{
		Filters: []types.Filter{{
			Name: aws.String("tag-key"), Values: []string{c.awsTag},
		}},
	})
	cache := make(map[string]string)
	for {
		if !p.HasMorePages() {
			break
		}
		out, err := p.NextPage(c.ctx)
		if err != nil {
			c.Error(err, "refreshing cache with paginator NextPage")
			return
		}

		for _, r := range out.Reservations {
			for _, i := range r.Instances {
				cache[*i.InstanceId] = ComputeNodeLabelKeySuffixForInstance(i)
			}
		}
	}
	c.Lock()
	c.cache = cache
	c.Unlock()
}

func (c *Cache) Get(instanceId string) (string, error) {
	c.RLock()
	role, ok := c.cache[instanceId]
	c.RUnlock()
	if !ok {
		instance, err := GetEC2InstanceById(c.ctx, instanceId)
		if err != nil {
			return "", err
		}
		role = ComputeNodeLabelKeySuffixForInstance(instance)
		c.Lock()
		c.cache[instanceId] = role
		c.Unlock()
	}
	return role, nil
}

func (c *Cache) Del(instanceId string) {
	c.Lock()
	delete(c.cache, instanceId)
	c.Unlock()
}
