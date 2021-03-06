package main

import (
	"os"
	"os/exec"
	"testing"

	. "github.com/franela/goblin"
)

func newBool(b bool) *bool {
	return &b
}

func TestPlugin(t *testing.T) {
	g := Goblin(t)

	g.Describe("CopyTfEnv", func() {
		g.It("Should create copies of TF_VAR_ to lowercase", func() {
			// Set some initial TF_VAR_ that are uppercase
			os.Setenv("TF_VAR_SOMETHING", "some value")
			os.Setenv("TF_VAR_SOMETHING_ELSE", "some other value")
			os.Setenv("TF_VAR_BASE64", "dGVzdA==")

			CopyTfEnv()

			// Make sure new env vars exist with proper values
			g.Assert(os.Getenv("TF_VAR_something")).Equal("some value")
			g.Assert(os.Getenv("TF_VAR_something_else")).Equal("some other value")
			g.Assert(os.Getenv("TF_VAR_base64")).Equal("dGVzdA==")
		})
	})

	g.Describe("tfApply", func() {
		g.It("Should return correct apply commands given the arguments", func() {
			type args struct {
				config Config
			}

			tests := []struct {
				name string
				args args
				want *exec.Cmd
			}{
				{
					"default",
					args{config: Config{}},
					exec.Command("terraform", "apply", "plan.tfout"),
				},
				{
					"with parallelism",
					args{config: Config{Parallelism: 5}},
					exec.Command("terraform", "apply", "-parallelism=5", "plan.tfout"),
				},
				{
					"with targets",
					args{config: Config{Targets: []string{"target1", "target2"}}},
					exec.Command("terraform", "apply", "--target", "target1", "--target", "target2", "plan.tfout"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfApply(tt.args.config)).Equal(tt.want)
			}
		})
	})

	g.Describe("tfDestroy", func() {
		g.It("Should return correct destroy commands given the arguments", func() {
			type args struct {
				config Config
			}

			tests := []struct {
				name string
				args args
				want *exec.Cmd
			}{
				{
					"default",
					args{config: Config{}},
					exec.Command("terraform", "destroy", "-force"),
				},
				{
					"with parallelism",
					args{config: Config{Parallelism: 5}},
					exec.Command("terraform", "destroy", "-parallelism=5", "-force"),
				},
				{
					"with targets",
					args{config: Config{Targets: []string{"target1", "target2"}}},
					exec.Command("terraform", "destroy", "-target=target1", "-target=target2", "-force"),
				},
				{
					"with vars",
					args{config: Config{Vars: map[string]string{"username": "someuser", "password": "1pass"}}},
					exec.Command("terraform", "destroy", "-var", "password=1pass", "-var", "username=someuser", "-force"),
				},
				{
					"with var-files",
					args{config: Config{VarFiles: []string{"common.tfvars", "prod.tfvars"}}},
					exec.Command("terraform", "destroy", "-var-file=common.tfvars", "-var-file=prod.tfvars", "-force"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfDestroy(tt.args.config)).Equal(tt.want)
			}
		})
	})

	g.Describe("tfImport", func() {
		g.It("Should return correct import commands given the arguments", func() {
			type args struct {
				config Config
				target string
				id string
			}

			tests := []struct {
				name string
				args args
				want *exec.Cmd
			}{
				{
					"default",
					args{config: Config{}, target: "foo", id: "bar"},
					exec.Command("terraform", "import", "foo", "bar"),
				},
				{
					"with lock",
					args{config: Config{InitOptions: InitOptions{Lock: newBool(true)}}, target: "foo", id: "bar"},
					exec.Command("terraform", "import", "-lock=true", "foo", "bar"),
				},
				{
					"with lock-timeout",
					args{config: Config{InitOptions: InitOptions{LockTimeout: "1s"}}, target: "foo", id: "bar"},
					exec.Command("terraform", "import", "-lock-timeout=1s", "foo", "bar"),
				},
				{
					"with vars",
					args{config: Config{Vars: map[string]string{"username": "someuser", "password": "1pass"}}, target: "foo", id: "bar"},
					exec.Command("terraform", "import", "-var", "password=1pass", "-var", "username=someuser", "foo", "bar"),
				},
				{
					"with var-files",
					args{config: Config{VarFiles: []string{"common.tfvars", "prod.tfvars"}}, target: "foo", id: "bar"},
					exec.Command("terraform", "import", "-var-file=common.tfvars", "-var-file=prod.tfvars", "foo", "bar"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfImport(tt.args.config, tt.args.target, tt.args.id)).Equal(tt.want)
			}
		})
	})

	g.Describe("tfPlan", func() {
		g.It("Should return correct plan commands given the arguments", func() {
			type args struct {
				config Config
			}

			tests := []struct {
				name    string
				args    args
				destroy bool
				want    *exec.Cmd
			}{
				{
					"default",
					args{config: Config{}},
					false,
					exec.Command("terraform", "plan", "-out=plan.tfout"),
				},
				{
					"destroy",
					args{config: Config{}},
					true,
					exec.Command("terraform", "plan", "-destroy"),
				},
				{
					"with vars",
					args{config: Config{Vars: map[string]string{"username": "someuser", "password": "1pass"}}},
					false,
					exec.Command("terraform", "plan", "-out=plan.tfout", "-var", "password=1pass", "-var", "username=someuser"),
				},
				{
					"with var-files",
					args{config: Config{VarFiles: []string{"common.tfvars", "prod.tfvars"}}},
					false,
					exec.Command("terraform", "plan", "-out=plan.tfout", "-var-file=common.tfvars", "-var-file=prod.tfvars"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfPlan(tt.args.config, tt.destroy)).Equal(tt.want)
			}
		})
	})
	g.Describe("tfFmt", func() {
		g.It("Should return correct fmt commands given the arguments", func() {
			type args struct {
				config Config
			}

			affirmative := true
			negative := false

			tests := []struct {
				name string
				args args
				want *exec.Cmd
			}{
				{
					"default",
					args{config: Config{}},
					exec.Command("terraform", "fmt"),
				},
				{
					"with list",
					args{config: Config{FmtOptions: FmtOptions{List: &affirmative}}},
					exec.Command("terraform", "fmt", "-list=true"),
				},
				{
					"with write",
					args{config: Config{FmtOptions: FmtOptions{Write: &affirmative}}},
					exec.Command("terraform", "fmt", "-write=true"),
				},
				{
					"with diff",
					args{config: Config{FmtOptions: FmtOptions{Diff: &affirmative}}},
					exec.Command("terraform", "fmt", "-diff=true"),
				},
				{
					"with check",
					args{config: Config{FmtOptions: FmtOptions{Check: &affirmative}}},
					exec.Command("terraform", "fmt", "-check=true"),
				},
				{
					"with combination",
					args{config: Config{FmtOptions: FmtOptions{
						List:  &negative,
						Write: &negative,
						Diff:  &affirmative,
						Check: &affirmative,
					}}},
					exec.Command("terraform", "fmt", "-list=false", "-write=false", "-diff=true", "-check=true"),
				},
			}

			for _, tt := range tests {
				g.Assert(tfFmt(tt.args.config)).Equal(tt.want)
			}
		})
	})
}
