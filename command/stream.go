// Licensed to Elasticsearch under one or more contributor
// license agreements. See the NOTICE file distributed with
// this work for additional information regarding copyright
// ownership. Elasticsearch licenses this file to you under
// the Apache License, Version 2.0 (the "License"); you may
// not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing,
// software distributed under the License is distributed on an
// "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY
// KIND, either express or implied.  See the License for the
// specific language governing permissions and limitations
// under the License.

package command

import (
	"flag"
	"github.com/elasticsearch/kriterium/panics"
	"lsf"
	"lsf/schema"
)

const cmd_stream lsf.CommandCode = "stream"

type streamOptionsSpec struct {
	verbose BoolOptionSpec
}

type editStreamOptionsSpec struct {
	verbose BoolOptionSpec
	global  BoolOptionSpec
	id      StringOptionSpec
	mode    StringOptionSpec
	path    StringOptionSpec
	pattern StringOptionSpec
}

func initEditStreamOptionsSpec(flagset *flag.FlagSet) *editStreamOptionsSpec {
	return &editStreamOptionsSpec{
		verbose: NewBoolFlag(flagset, "v", "verbose", false, "be verbose in list", false),
		global:  NewBoolFlag(flagset, "G", "global", false, "global scope flag for command", false),
		id:      NewStringFlag(flagset, "s", "stream-id", "", "unique identifier for stream", true),
		path:    NewStringFlag(flagset, "p", "path", "", "path to log files", true),
		mode:    NewStringFlag(flagset, "m", "journal-mode", "", "stream journaling mode (rotation|rollover)", false),
		pattern: NewStringFlag(flagset, "n", "name-pattern", "", "naming pattern of journaled log files", true),
	}
}
func _verifyEditStreamRequiredOpts(env *lsf.Environment, args ...string) (err error) {
	//	defer panics.Recover(&err)

	options := []interface{}{
		addStreamOptions.id,
		addStreamOptions.pattern,
		addStreamOptions.path,
	}
	e := verifyRequiredOptions(options)
	panics.OnError(e, "usage")

	mode := *addStreamOptions.mode.value
	switch schema.ToJournalModel(mode) {
	case schema.JournalModel.Rotation, schema.JournalModel.Rollover: // OK
	default: // not OK
		panics.OnFalse(false, "stream-add", "option", "option mode must be one of {rollover, rotation}")
	}
	return
}

var Stream *lsf.Command
var streamOptions *streamOptionsSpec

const (
	streamOptionVerbose   = "command.stream.option.verbose"
	streamOptionGlobal    = "command.stream.option.global"
	streamOptionsSelected = "command.stream.option.selected"
)

func init() {

	Stream = &lsf.Command{
		Name:  cmd_stream,
		About: "Stream is a top level command for log stream configuration and management",
		Run:   runStream,
		Flag:  FlagSet(cmd_stream),
	}
	streamOptions = &streamOptionsSpec{
		verbose: NewBoolFlag(Stream.Flag, "v", "verbose", false, "be verbose in list", false),
	}
}

func runStream(env *lsf.Environment, args ...string) error {

	if *streamOptions.verbose.value {
		env.Set(streamOptionVerbose, true)
	}

	xoff := 0
	var subcmd *lsf.Command = listStream
	if len(args) > 0 {
		subcmd = getSubCommand(args[0])
		xoff = 1
	}

	return lsf.Run(env, subcmd, args[xoff:]...)
}

func getSubCommand(subcmd string) *lsf.Command {

	var cmd *lsf.Command
	switch lsf.CommandCode("stream-" + subcmd) {
	case addStreamCmdCode:
		cmd = addStream
	case removeStreamCmdCode:
		cmd = removeStream
	case updateStreamCmdCode:
		cmd = updateStream
	case listStreamCmdCode:
		cmd = listStream
	default:
		// not panic -- return error TODO
		panic("BUG - unknown subcommand for stream: " + subcmd)
	}
	return cmd
}
