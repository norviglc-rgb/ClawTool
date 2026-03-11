package remote

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/openclaw/clawtool/internal/core"
	"golang.org/x/crypto/ssh"
	"golang.org/x/crypto/ssh/knownhosts"
)

// Executor runs one command on a resolved remote target. / Executor 在解析后的远程目标上执行一条命令。
type Executor interface {
	Execute(ctx context.Context, options ConnectionOptions, command string) (ExecOutput, error)
}

// ConnectionOptions stores resolved SSH connection settings. / ConnectionOptions 保存解析后的 SSH 连接设置。
type ConnectionOptions struct {
	Host            string
	Port            int
	User            string
	KeyPath         string
	HostKeyStrategy string
	KnownHostsPath  string
	OriginalAddress string
}

// ExecOutput stores deterministic remote process output. / ExecOutput 保存确定性的远程进程输出。
type ExecOutput struct {
	Stdout   string
	Stderr   string
	ExitCode int
	Duration time.Duration
}

// SSHExecutor executes commands over native Go SSH transport. / SSHExecutor 使用 Go 原生 SSH 传输执行命令。
type SSHExecutor struct{}

// NewSSHExecutor creates the default SSH executor. / NewSSHExecutor 创建默认 SSH 执行器。
func NewSSHExecutor() SSHExecutor {
	return SSHExecutor{}
}

// Execute connects to the remote host and runs the provided command. / Execute 连接远程主机并执行给定命令。
func (SSHExecutor) Execute(ctx context.Context, options ConnectionOptions, command string) (ExecOutput, error) {
	signer, err := loadPrivateKey(options.KeyPath)
	if err != nil {
		return ExecOutput{}, err
	}

	callback, err := hostKeyCallback(options)
	if err != nil {
		return ExecOutput{}, err
	}

	config := &ssh.ClientConfig{
		User:            options.User,
		Auth:            []ssh.AuthMethod{ssh.PublicKeys(signer)},
		HostKeyCallback: callback,
		Timeout:         10 * time.Second,
	}

	address := net.JoinHostPort(options.Host, strconv.Itoa(options.Port))
	dialer := &net.Dialer{Timeout: 10 * time.Second}
	conn, err := dialer.DialContext(ctx, "tcp", address)
	if err != nil {
		return ExecOutput{}, &core.AppError{
			Code:       core.ErrorCodeRemoteExec,
			MessageKey: "error.remote.exec",
			Cause:      err,
		}
	}

	clientConn, chans, reqs, err := ssh.NewClientConn(conn, address, config)
	if err != nil {
		_ = conn.Close()
		return ExecOutput{}, &core.AppError{
			Code:       core.ErrorCodeRemoteExec,
			MessageKey: "error.remote.exec",
			Cause:      err,
		}
	}
	client := ssh.NewClient(clientConn, chans, reqs)
	defer client.Close()

	session, err := client.NewSession()
	if err != nil {
		return ExecOutput{}, &core.AppError{
			Code:       core.ErrorCodeRemoteExec,
			MessageKey: "error.remote.exec",
			Cause:      err,
		}
	}
	defer session.Close()

	var stdout bytes.Buffer
	var stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			_ = client.Close()
		case <-done:
		}
	}()
	startedAt := time.Now()
	runErr := session.Run(command)
	close(done)

	output := ExecOutput{
		Stdout:   strings.TrimSpace(stdout.String()),
		Stderr:   strings.TrimSpace(stderr.String()),
		ExitCode: 0,
		Duration: time.Since(startedAt),
	}

	if runErr == nil {
		return output, nil
	}

	var exitErr *ssh.ExitError
	if errors.As(runErr, &exitErr) {
		output.ExitCode = exitErr.ExitStatus()
		return output, nil
	}
	if errors.Is(ctx.Err(), context.Canceled) || errors.Is(ctx.Err(), context.DeadlineExceeded) {
		return output, &core.AppError{
			Code:       core.ErrorCodeRemoteExec,
			MessageKey: "error.remote.exec",
			Cause:      ctx.Err(),
		}
	}

	return output, &core.AppError{
		Code:       core.ErrorCodeRemoteExec,
		MessageKey: "error.remote.exec",
		Cause:      runErr,
	}
}

func loadPrivateKey(path string) (ssh.Signer, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, &core.AppError{
			Code:       core.ErrorCodeRemoteExec,
			MessageKey: "error.remote.exec",
			Cause:      err,
		}
	}

	signer, err := ssh.ParsePrivateKey(data)
	if err != nil {
		return nil, &core.AppError{
			Code:       core.ErrorCodeRemoteExec,
			MessageKey: "error.remote.exec",
			Cause:      err,
		}
	}
	return signer, nil
}

func hostKeyCallback(options ConnectionOptions) (ssh.HostKeyCallback, error) {
	switch options.HostKeyStrategy {
	case "insecure":
		return ssh.InsecureIgnoreHostKey(), nil
	case "", "known_hosts":
		callback, err := knownhosts.New(options.KnownHostsPath)
		if err != nil {
			return nil, &core.AppError{
				Code:       core.ErrorCodeRemoteExec,
				MessageKey: "error.remote.exec",
				Cause:      err,
			}
		}
		return callback, nil
	default:
		return nil, &core.AppError{
			Code:       core.ErrorCodeRemoteExec,
			MessageKey: "error.remote.exec",
			Cause:      fmt.Errorf("unsupported host key strategy: %s", options.HostKeyStrategy),
		}
	}
}

func resolveConnectionOptions(profile core.Profile) (ConnectionOptions, error) {
	if profile.Target.Kind != "ssh" {
		return ConnectionOptions{}, &core.AppError{
			Code:       core.ErrorCodeRemoteExec,
			MessageKey: "error.remote.exec",
			Cause:      fmt.Errorf("profile %s is not an ssh target", profile.Name),
		}
	}

	address := strings.TrimSpace(profile.Target.Address)
	if address == "" {
		return ConnectionOptions{}, &core.AppError{
			Code:       core.ErrorCodeRemoteExec,
			MessageKey: "error.remote.exec",
			Cause:      fmt.Errorf("profile %s is missing ssh address", profile.Name),
		}
	}

	parsedUser, parsedHost := splitSSHAddress(address)
	host := parsedHost
	if strings.TrimSpace(host) == "" {
		host = address
	}

	currentUser, _ := user.Current()
	resolvedUser := strings.TrimSpace(profile.Target.User)
	if resolvedUser == "" {
		resolvedUser = parsedUser
	}
	if resolvedUser == "" && currentUser != nil {
		resolvedUser = currentUser.Username
	}
	if strings.TrimSpace(resolvedUser) == "" {
		return ConnectionOptions{}, &core.AppError{
			Code:       core.ErrorCodeRemoteExec,
			MessageKey: "error.remote.exec",
			Cause:      fmt.Errorf("unable to resolve ssh user"),
		}
	}

	port := profile.Target.Port
	if port == 0 {
		port = 22
	}

	hostKeyStrategy := strings.TrimSpace(profile.Target.HostKeyStrategy)
	if hostKeyStrategy == "" {
		hostKeyStrategy = "known_hosts"
	}

	keyPath := strings.TrimSpace(profile.Target.KeyPath)
	if keyPath == "" {
		keyPath = detectDefaultKeyPath()
	}
	if keyPath == "" {
		return ConnectionOptions{}, &core.AppError{
			Code:       core.ErrorCodeRemoteExec,
			MessageKey: "error.remote.exec",
			Cause:      fmt.Errorf("unable to resolve ssh private key"),
		}
	}

	knownHostsPath := defaultKnownHostsPath()
	return ConnectionOptions{
		Host:            host,
		Port:            port,
		User:            resolvedUser,
		KeyPath:         keyPath,
		HostKeyStrategy: hostKeyStrategy,
		KnownHostsPath:  knownHostsPath,
		OriginalAddress: address,
	}, nil
}

func splitSSHAddress(value string) (string, string) {
	parts := strings.SplitN(value, "@", 2)
	if len(parts) == 2 {
		return strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	}
	return "", strings.TrimSpace(value)
}

func detectDefaultKeyPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}

	candidates := []string{
		filepath.Join(home, ".ssh", "id_ed25519"),
		filepath.Join(home, ".ssh", "id_ecdsa"),
		filepath.Join(home, ".ssh", "id_rsa"),
	}
	for _, candidate := range candidates {
		if _, err := os.Stat(candidate); err == nil {
			return candidate
		}
	}
	return ""
}

func defaultKnownHostsPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return filepath.Join(home, ".ssh", "known_hosts")
}
