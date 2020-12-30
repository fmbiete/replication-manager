// replication-manager - Replication Manager Monitoring and CLI for MariaDB and MySQL
// Copyright 2017 Signal 18 SARL
// Authors: Guillaume Lefranc <guillaume@signal18.io>
//          Stephane Varoqui  <svaroqui@gmail.com>
// This source code is licensed under the GNU General Public License, version 3.

package cluster

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"time"
)

func (cluster *Cluster) LocalhostUnprovisionProxySQLService(prx *Proxy) error {
	cluster.LocalhostStopProxysqlService(prx)
	cluster.errorChan <- nil
	return nil
}

func (cluster *Cluster) LocalhostProvisionProxySQLService(prx *Proxy) error {

	out := &bytes.Buffer{}
	path := prx.Datadir + "/var"
	//os.RemoveAll(path)

	cmd := exec.Command("rm", "-rf", path)

	cmd.Stdout = out
	err := cmd.Run()
	if err != nil {
		cluster.LogPrintf(LvlErr, "%s", err)
		cluster.errorChan <- err
		return err
	}
	cluster.LogPrintf(LvlInfo, "Remove datadir done: %s", out.Bytes())
	prx.GetProxyConfig()
	os.Symlink(prx.Datadir+"/init/data", path)

	err = cluster.LocalhostStartProxySQLService(prx)
	if err != nil {
		cluster.errorChan <- err
		return err

	}
	cluster.errorChan <- nil
	return nil
}

func (cluster *Cluster) LocalhostStopProxysqlService(prx *Proxy) error {

	//	cluster.LogPrintf("TEST", "Killing database %s %d", server.Id, server.Process.Pid)

	killCmd := exec.Command("kill", "-9", fmt.Sprintf("%d", prx.Process.Pid))
	killCmd.Run()
	return nil
}

func (cluster *Cluster) LocalhostStartProxySQLService(prx *Proxy) error {
	prx.GetProxyConfig()

	/*	path := prx.Datadir + "/var"
			err := os.RemoveAll(path + "/" + server.Id + ".pid")
				if err != nil {
					cluster.LogPrintf(LvlErr, "%s", err)
					return err
				}
			usr, err := user.Current()
		if err != nil {
			cluster.LogPrintf(LvlErr, "%s", err)
			return err
		}	*/

	mariadbdCmd := exec.Command(cluster.Conf.ProxysqlBinaryPath, "--config="+prx.Datadir+"/init/conf/proxysql.cnf", "--datadir="+prx.Datadir+"/var")
	cluster.LogPrintf(LvlInfo, "%s %s", mariadbdCmd.Path, mariadbdCmd.Args)

	var out bytes.Buffer
	mariadbdCmd.Stdout = &out

	go func() {
		err := mariadbdCmd.Run()
		if err != nil {
			cluster.LogPrintf(LvlErr, "%s ", err)
			fmt.Printf("Command finished with error: %v", err)
		}
	}()
	time.Sleep(time.Millisecond * 2000)
	prx.Process = mariadbdCmd.Process
	//	mariadbdCmd.Process.Release()

	return nil
}
