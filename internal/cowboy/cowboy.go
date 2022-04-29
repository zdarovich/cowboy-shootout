package cowboy

import (
	"context"
	pbCowboy "github.com/zdarovich/cowboy_shooters/api/cowboy"
	pbLog "github.com/zdarovich/cowboy_shooters/api/log"
	"github.com/zdarovich/cowboy_shooters/internal/util"
	"fmt"
	"github.com/pkg/errors"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	v1 "k8s.io/api/core/v1"
	kubeErrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"log"
	"math/rand"
	"strconv"
	"sync"
	"time"
)

const (
	DefaultNamespace = "default"
)

type Service struct {
	mu         sync.RWMutex
	k8sClient  *kubernetes.Clientset
	ctx        context.Context
	cancel     context.CancelFunc
	randSrc    *rand.Rand
	Name       string
	health     int
	Damage     int
	logSvcAddr string
}

func NewCowboy(name, health, damage, logSvcAddr string) (*Service, error) {
	config, err := rest.InClusterConfig()
	if err != nil {
		return nil, err
	}
	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, err
	}
	healthInt, err := strconv.Atoi(health)
	if err != nil {
		return nil, err
	}
	damageInt, err := strconv.Atoi(damage)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		health:     healthInt,
		k8sClient:  clientset,
		ctx:        ctx,
		cancel:     cancel,
		randSrc:    rand.New(rand.NewSource(time.Now().UnixNano())),
		logSvcAddr: logSvcAddr,
		Damage:     damageInt,
		Name:       name,
	}, nil
}

func (s *Service) Run() {
	err := s.log(fmt.Sprintf("%s with %d health and %d damage was initialized", s.Name, s.health, s.Damage))
	if err != nil {
		log.Printf("%v \n", err)
	}
	util.Until(func() {
		for {
			if s.GetHealth() <= 0 {
				err := s.log(fmt.Sprintf("%s realized that he is dead", s.Name))
				if err != nil {
					log.Printf("%v \n", err)
				}
				s.cancel()
			}

			var podList *v1.PodList
			getPods := func() error {
				podList, err = s.k8sClient.CoreV1().Pods(DefaultNamespace).List(s.ctx, metav1.ListOptions{
					LabelSelector: fmt.Sprintf("app=cowboy,name!=%s", s.Name),
					FieldSelector: "status.phase=Running",
				})
				if err != nil {
					return err
				}
				if len(podList.Items) == 0 {
					return errors.New("failed to get cowboys from k8s api")
				}
				if podList.Items[0].Status.PodIP == "" {
					return errors.New("failed to get cowboy ip from k8s api")
				}
				if len(podList.Items[0].Spec.Containers) == 0 ||
					len(podList.Items[0].Spec.Containers[0].Ports) == 0 {
					return errors.New("failed to get cowboy container port from k8s api")
				}
				return nil
			}
			err := util.Retry(5, time.Second, getPods)

			if kubeErrors.IsNotFound(err) || (podList != nil && len(podList.Items) == 0) {
				err := s.log(fmt.Sprintf("no alive cowboys. I am %s the winner!", s.Name))
				if err != nil {
					log.Printf("%v \n", err)
					continue
				}
				s.cancel()
				return
			}
			itemIdx := s.randSrc.Intn(len(podList.Items))
			target := fmt.Sprintf("%s:%d", podList.Items[itemIdx].Status.PodIP, podList.Items[itemIdx].Spec.Containers[0].Ports[0].ContainerPort)
			log.Printf("got target pod %s \n", target)
			msg, err := s.Shoot(target)
			if err != nil {
				log.Printf("failed to shoot cowbow: %v \n", err)
				continue
			}
			err = s.log(fmt.Sprintf("cowboy %s made a shoot and realized that %s", s.Name, msg))
			if err != nil {
				log.Printf("%v \n", err)
			}
			log.Printf("sleep for 1 sec\n")
			time.Sleep(time.Second)
		}
	}, time.Second, s.ctx.Done())
}

func (s *Service) Shoot(target string) (string, error) {
	conn, err := grpc.Dial(target, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return "", err
	}
	defer conn.Close()
	c := pbCowboy.NewCowboyClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(s.ctx, time.Second)
	defer cancel()

	resp, err := c.HandleShoot(ctx, &pbCowboy.ShootRequest{Damage: uint64(s.Damage)})
	if err != nil {
		return "", err
	}
	return resp.Msg, nil
}

func (s *Service) GetShoot(damage int) {
	s.mu.Lock()
	defer s.mu.Unlock()
	actualHealth := s.health - damage
	if actualHealth < 0 {
		actualHealth = 0
	}
	s.health = actualHealth
}

func (s *Service) GetHealth() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.health
}

func (s *Service) IsDead() <-chan struct{} {
	return s.ctx.Done()
}

func (s *Service) log(msg string) error {
	log.Println(msg)
	conn, err := grpc.Dial(s.logSvcAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return err
	}
	defer conn.Close()
	c := pbLog.NewLogClient(conn)

	// Contact the server and print out its response.
	ctx, cancel := context.WithTimeout(s.ctx, time.Second)
	defer cancel()

	_, err = c.HandleProduce(ctx, &pbLog.ProduceRequest{Record: &pbLog.Record{
		Value: []byte(msg),
	}})
	if err != nil {
		return err
	}
	return nil
}
