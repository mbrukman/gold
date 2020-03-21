package ppo

import (
	agentv1 "github.com/aunum/gold/pkg/v1/agent"
	envv1 "github.com/aunum/gold/pkg/v1/env"
	modelv1 "github.com/aunum/gold/pkg/v1/model"
	l "github.com/aunum/gold/pkg/v1/model/layers"
	"github.com/aunum/gold/pkg/v1/model/layers/activation"
	"github.com/aunum/gold/pkg/v1/model/layers/fc"
	"github.com/aunum/log"
	g "gorgonia.org/gorgonia"
)

// ModelConfig are the hyperparameters for a model.
type ModelConfig struct {
	// Loss function to evaluate network performance.
	Loss modelv1.Loss

	// Optimizer to optimize the weights with regards to the error.
	Optimizer g.Solver

	// LayerBuilder is a builder of layer.
	LayerBuilder LayerBuilder

	// BatchSize of the updates.
	BatchSize int

	// Track is whether to track the model.
	Track bool
}

// LayerBuilder builds layers.
type LayerBuilder func(env *envv1.Env) []l.Layer

// DefaultActorConfig are the default hyperparameters for a policy.
var DefaultActorConfig = &ModelConfig{
	Optimizer:    g.NewAdamSolver(),
	LayerBuilder: DefaultActorLayerBuilder,
	BatchSize:    20,
}

// DefaultActorLayerBuilder is a default fully connected layer builder.
var DefaultActorLayerBuilder = func(env *envv1.Env) []l.Layer {
	return []l.Layer{
		fc.New(env.ObservationSpaceShape()[0], 24, fc.WithName("fc1")),
		fc.New(24, 24, fc.WithName("fc2")),
		fc.New(24, envv1.PotentialsShape(env.ActionSpace)[0], fc.WithActivation(activation.Softmax), fc.WithName("predictions")),
	}
}

// MakeActor makes the actor which chooses actions based on the policy.
func MakeActor(config *ModelConfig, base *agentv1.Base, env *envv1.Env) (modelv1.Model, error) {
	x := modelv1.NewInput("x", []int{1, env.ObservationSpaceShape()[0]})
	y := modelv1.NewInput("y", []int{1, envv1.PotentialsShape(env.ActionSpace)[0]})

	oldPolicyProbs := modelv1.NewInput("oldProbs", []int{1, envv1.PotentialsShape(env.ActionSpace)[0]})
	advantages := modelv1.NewInput("advantages", []int{1, 1})
	rewards := modelv1.NewInput("rewards", []int{1, 1})
	values := modelv1.NewInput("values", []int{1, 1})

	log.Debugv("x shape", x.Shape())
	log.Debugv("y shape", y.Shape())

	model, err := modelv1.NewSequential("actor")
	if err != nil {
		return nil, err
	}
	model.AddLayers(config.LayerBuilder(env)...)

	loss := NewLoss(oldPolicyProbs, advantages, rewards, values)

	opts := modelv1.NewOpts()
	opts.Add(
		modelv1.WithOptimizer(config.Optimizer),
		modelv1.WithLoss(loss),
		modelv1.WithBatchSize(config.BatchSize),
		modelv1.WithTracker(base.Tracker),
		modelv1.WithMetrics(modelv1.TrainBatchLossMetric),
	)

	model.Fwd(x)
	err = model.Compile(
		modelv1.Inputs{x, oldPolicyProbs, advantages, rewards, values},
		y,
		opts.Values()...,
	)
	if err != nil {
		return nil, err
	}
	return model, nil
}

// DefaultCriticLayerBuilder is a default fully connected layer builder.
var DefaultCriticLayerBuilder = func(env *envv1.Env) []l.Layer {
	return []l.Layer{
		fc.New(env.ObservationSpaceShape()[0], 24, fc.WithName("w0")),
		fc.New(24, 24, fc.WithName("w1")),
		fc.New(24, 1, fc.WithActivation(activation.Tanh), fc.WithName("w2")),
	}
}

// DefaultCriticConfig are the default hyperparameters for a policy.
var DefaultCriticConfig = &ModelConfig{
	Loss:         modelv1.MSE,
	Optimizer:    g.NewAdamSolver(),
	LayerBuilder: DefaultCriticLayerBuilder,
	BatchSize:    20,
}

// MakeCritic makes the critic which creats a qValue based on the outcome of the action taken.
func MakeCritic(config *ModelConfig, base *agentv1.Base, env *envv1.Env) (modelv1.Model, error) {
	x := modelv1.NewInput("x", []int{1, env.ObservationSpaceShape()[0]})
	y := modelv1.NewInput("y", []int{1, 1})

	log.Debugv("x shape", x.Shape())
	log.Debugv("y shape", y.Shape())

	model, err := modelv1.NewSequential("critic")
	if err != nil {
		return nil, err
	}
	model.AddLayers(config.LayerBuilder(env)...)

	opts := modelv1.NewOpts()
	opts.Add(
		modelv1.WithOptimizer(config.Optimizer),
		modelv1.WithLoss(config.Loss),
		modelv1.WithBatchSize(config.BatchSize),
		modelv1.WithTracker(base.Tracker),
		modelv1.WithMetrics(modelv1.TrainBatchLossMetric),
	)

	model.Fwd(x)
	err = model.Compile(
		x,
		y,
		opts.Values()...,
	)
	if err != nil {
		return nil, err
	}
	return model, nil
}
