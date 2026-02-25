package main

import (
	"context"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/clientcmd"
)

var clientset *kubernetes.Clientset

func initClient() error {
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		homeDir := os.Getenv("HOME")
		kubeconfig = homeDir + "/.kube/config"
	}

	config, err := clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return fmt.Errorf("failed to load kubeconfig: %v", err)
	}

	clientset, err = kubernetes.NewForConfig(config)
	if err != nil {
		return fmt.Errorf("failed to create clientset: %v", err)
	}

	return nil
}

var rootCmd = &cobra.Command{
	Use:   "k8s-manager",
	Short: "Kubernetes Cluster Management CLI",
	PersistentPreRun: func(cmd *cobra.Command, args []string) {
		if err := initClient(); err != nil {
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			os.Exit(1)
		}
	},
}

var nodesCmd = &cobra.Command{
	Use:   "nodes",
	Short: "List all nodes",
	Run: func(cmd *cobra.Command, args []string) {
		nodes, err := clientset.CoreV1().Nodes().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get nodes: %v\n", err)
			return
		}

		fmt.Printf("%-20s %-15s %-20s %-15s\n", "NAME", "STATUS", "ROLES", "AGE")
		fmt.Println("----------------------------------------------------------------")
		for _, node := range nodes.Items {
			age := time.Since(node.CreationTimestamp.Time).Round(time.Hour)
			roles := ""
			for k, v := range node.Labels {
				if v == "master" || v == "control-plane" {
					roles += k + ","
				}
			}
			if roles == "" {
				roles = "worker"
			}
			fmt.Printf("%-20s %-15s %-20s %-15v\n", 
				node.Name, 
				node.Status.Conditions[len(node.Status.Conditions)-1].Type,
				roles,
				age,
			)
		}
	},
}

var podsCmd = &cobra.Command{
	Use:   "pods [namespace]",
	Short: "List pods in a namespace",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ns := "default"
		if len(args) > 0 {
			ns = args[0]
		}

		pods, err := clientset.CoreV1().Pods(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get pods: %v\n", err)
			return
		}

		fmt.Printf("%-30s %-15s %-10s %-15s\n", "NAME", "READY", "STATUS", "AGE")
		fmt.Println("----------------------------------------------------------------")
		for _, pod := range pods.Items {
			age := time.Since(pod.CreationTimestamp.Time).Round(time.Hour)
			ready := fmt.Sprintf("%d/%d", getReadyContainers(&pod), len(pod.Spec.Containers))
			fmt.Printf("%-30s %-15s %-10s %-15v\n", 
				pod.Name, 
				ready, 
				string(pod.Status.Phase),
				age,
			)
		}
	},
}

func getReadyContainers(pod *corev1.Pod) int {
	ready := 0
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
	}
	return ready
}

var deployCmd = &cobra.Command{
	Use:   "deploy [namespace]",
	Short: "List deployments in a namespace",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ns := "default"
		if len(args) > 0 {
			ns = args[0]
		}

		deploys, err := clientset.AppsV1().Deployments(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get deployments: %v\n", err)
			return
		}

		fmt.Printf("%-25s %-10s %-15s %-15s %-15s\n", "NAME", "READY", "UP-TO-DATE", "AVAILABLE", "AGE")
		fmt.Println("----------------------------------------------------------------")
		for _, d := range deploys.Items {
			age := time.Since(d.CreationTimestamp.Time).Round(time.Hour)
			ready := fmt.Sprintf("%d/%d", d.Status.ReadyReplicas, d.Status.Replicas)
			fmt.Printf("%-25s %-10s %-15d %-15d %-15v\n", 
				d.Name, 
				ready,
				d.Status.UpdatedReplicas,
				d.Status.AvailableReplicas,
				age,
			)
		}
	},
}

var scaleCmd = &cobra.Command{
	Use:   "scale <deployment> <replicas> [namespace]",
	Short: "Scale a deployment",
	Args:  cobra.MinimumNArgs(2),
	Run: func(cmd *cobra.Command, args []string) {
		deployName := args[0]
		replicas := int32(0)
		fmt.Sscanf(args[1], "%d", &replicas)
		ns := "default"
		if len(args) > 2 {
			ns = args[2]
		}

		deploy, err := clientset.AppsV1().Deployments(ns).Get(context.Background(), deployName, metav1.GetOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get deployment: %v\n", err)
			return
		}

		deploy.Spec.Replicas = &replicas
		_, err = clientset.AppsV1().Deployments(ns).Update(context.Background(), deploy, metav1.UpdateOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to scale deployment: %v\n", err)
			return
		}

		fmt.Printf("✓ Scaled deployment %s to %d replicas\n", deployName, replicas)
	},
}

var logsCmd = &cobra.Command{
	Use:   "logs <pod> [namespace]",
	Short: "Get logs from a pod",
	Args:  cobra.RangeArgs(1, 2),
	Run: func(cmd *cobra.Command, args []string) {
		podName := args[0]
		ns := "default"
		if len(args) > 1 {
			ns = args[1]
		}

		tailLines := int64(100)
		logs, err := clientset.CoreV1().Pods(ns).GetLogs(podName, &corev1.PodLogOptions{TailLines: &tailLines}).Do(context.Background()).Raw()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get logs: %v\n", err)
			return
		}

		fmt.Println(string(logs))
	},
}

var svcCmd = &cobra.Command{
	Use:   "svc [namespace]",
	Short: "List services in a namespace",
	Args:  cobra.MaximumNArgs(1),
	Run: func(cmd *cobra.Command, args []string) {
		ns := "default"
		if len(args) > 0 {
			ns = args[0]
		}

		svcs, err := clientset.CoreV1().Services(ns).List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get services: %v\n", err)
			return
		}

		fmt.Printf("%-25s %-20s %-15s %-15s\n", "NAME", "TYPE", "CLUSTER-IP", "PORT(S)")
		fmt.Println("----------------------------------------------------------------")
		for _, s := range svcs.Items {
			ports := ""
			for _, p := range s.Spec.Ports {
				ports += fmt.Sprintf("%d/%s,", p.Port, p.Protocol)
			}
			if len(ports) > 0 {
				ports = ports[:len(ports)-1]
			}
			fmt.Printf("%-25s %-20s %-15s %-15s\n", 
				s.Name, 
				string(s.Spec.Type),
				s.Spec.ClusterIP,
				ports,
			)
		}
	},
}

var nsCmd = &cobra.Command{
	Use:   "ns",
	Short: "List all namespaces",
	Run: func(cmd *cobra.Command, args []string) {
		nss, err := clientset.CoreV1().Namespaces().List(context.Background(), metav1.ListOptions{})
		if err != nil {
			fmt.Fprintf(os.Stderr, "Failed to get namespaces: %v\n", err)
			return
		}

		fmt.Printf("%-20s %-15s\n", "NAME", "STATUS")
		fmt.Println("----------------------------------------------------------------")
		for _, ns := range nss.Items {
			fmt.Printf("%-20s %-15s\n", ns.Name, string(ns.Status.Phase))
		}
	},
}

func main() {
	rootCmd.AddCommand(nodesCmd)
	rootCmd.AddCommand(podsCmd)
	rootCmd.AddCommand(deployCmd)
	rootCmd.AddCommand(scaleCmd)
	rootCmd.AddCommand(logsCmd)
	rootCmd.AddCommand(svcCmd)
	rootCmd.AddCommand(nsCmd)

	if err := rootCmd.Execute(); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
