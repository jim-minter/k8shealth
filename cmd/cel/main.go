package main

import (
	"fmt"
	"log"

	"github.com/google/cel-go/cel"
	"google.golang.org/protobuf/protoadapt"
	corev1 "k8s.io/api/core/v1"
)

// example usages of cel evaluation in strongly and weakly typed modes

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	{
		// strongly typed example
		pod := &corev1.Pod{
			Status: corev1.PodStatus{
				Conditions: []corev1.PodCondition{
					{
						Type:   "test",
						Status: corev1.ConditionTrue,
						Reason: "all good",
					},
				},
			},
		}

		env, err := cel.NewEnv(
			cel.Types(protoadapt.MessageV2Of(&corev1.Pod{})),
			cel.Variable("pod", cel.ObjectType("k8s.io.api.core.v1.Pod")),
		)
		if err != nil {
			return err
		}

		_, issues := env.Compile(`pod.doesntexist == "True"`)
		if issues.Err() != nil {
			fmt.Println(issues.Err())
		}

		cel, issues := env.Compile(`pod.status.conditions.filter(c, c.type == "test" && c.status == "True")[0].reason`)
		if issues.Err() != nil {
			return issues.Err()
		}

		program, err := env.Program(cel)
		if err != nil {
			return err
		}

		out, _, err := program.Eval(map[string]any{
			"pod": protoadapt.MessageV2Of(pod),
		})
		if err != nil {
			return err
		}

		fmt.Println(out)
	}

	{
		// dynamic typing example
		dynamic := map[string]any{
			"number": float64(1),
			"string": "str",
		}

		env, err := cel.NewEnv(
			cel.Variable("dynamic", cel.DynType),
		)
		if err != nil {
			return err
		}

		cel, issues := env.Compile(`dynamic.number`)
		if issues.Err() != nil {
			return issues.Err()
		}

		program, err := env.Program(cel)
		if err != nil {
			return err
		}

		out, _, err := program.Eval(map[string]any{
			"dynamic": dynamic,
		})
		if err != nil {
			return err
		}

		fmt.Println(out)
	}

	return nil
}
