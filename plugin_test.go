package main

import (
	"io/ioutil"
	"testing"
	"github.com/kadende/kadende-interfaces/spi/instance"
	"github.com/kadende/kadende-interfaces/pkg/types"
	"github.com/stretchr/testify/require"
	"path/filepath"
	"os"
)

func TestUsage(t *testing.T) {
	run(t, `
{
            "ami" : "${lookup(var.aws_amis, var.aws_region)}",
            "instance_type" : "m1.small",
            "key_name": "PUBKEY",
            "vpc_security_group_ids" : ["${aws_security_group.default.id}"],
            "subnet_id": "${aws_subnet.default.id}"
}`)
}

func run(t *testing.T, properties string) {
	dir, err := ioutil.TempDir(".", "kadende-instance-file")
	require.NoError(t, err)

	defer os.RemoveAll(dir)

	fileinst := NewPlugin(dir)

	config := types.AnyString(properties)

	err = fileinst.Validate(config)
	require.NoError(t, err)

	instanceSpec := instance.Spec{
		Properties: config,
		Tags: map[string]string{
			"label1": "value1",
			"label2": "value2",
		},
		Init: "apt-get update -y\n\napt-get install -y software",
		Attachments: []instance.Attachment{
			{
				ID:   "ebs1",
				Type: "ebs",
			},
		},
	}

	id, err := fileinst.Provision(instanceSpec)
	require.NoError(t, err)

	tfPath := filepath.Join(dir, string(*id))
	buff, err := ioutil.ReadFile(tfPath)
	require.NoError(t, err)

	any := types.AnyBytes(buff)
	parsed := fileInstance{}
	err = any.Decode(&parsed)
	require.NoError(t, err)
	require.Equal(t, map[string]string{
		"label1": "value1",
		"label2": "value2",
	}, parsed.Spec.Tags)
	require.Equal(t, instanceSpec.Init, parsed.Spec.Init)

	// label resources
	err = fileinst.Label(*id, map[string]string{
		"label1": "changed1",
		"label3": "value3",
	})

	buff, err = ioutil.ReadFile(tfPath)
	require.NoError(t, err)

	any = types.AnyBytes(buff)
	parsed = fileInstance{}
	err = any.Decode(&parsed)
	require.NoError(t, err)

	require.Equal(t, map[string]string{
		"label1": "changed1",
		"label2": "value2",
		"label3": "value3",
	}, parsed.Spec.Tags)

	list, err := fileinst.DescribeInstances(map[string]string{"label1": "changed1"}, false)
	require.NoError(t, err)
	require.Equal(t, []instance.Description{
		{
			ID:   *id,
			Tags: parsed.Spec.Tags,
		},
	}, list)

	err = fileinst.Destroy(*id, instance.Termination)
	require.NoError(t, err)

	if _, err := os.Stat(tfPath); !os.IsNotExist(err) {
		require.Fail(t,"File still exists")
	}

	list, err = fileinst.DescribeInstances(map[string]string{"label1": "changed1"}, false)
	require.NoError(t, err)
	require.Equal(t, []instance.Description{}, list)

}