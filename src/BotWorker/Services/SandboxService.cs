using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;
using System.Text;
using System.Threading;
using System.Threading.Tasks;
using Docker.DotNet;
using Docker.DotNet.Models;
using Microsoft.Extensions.Logging;

namespace BotWorker.Services
{
    public class Sandbox
    {
        public string ID { get; set; } = string.Empty;
        private readonly SandboxService _service;

        public Sandbox(string id, SandboxService service)
        {
            ID = id;
            _service = service;
        }

        public async Task<(string stdout, string stderr)> ExecAsync(string command, CancellationToken ct = default)
        {
            return await _service.ExecInContainerAsync(ID, command, ct);
        }

        public async Task WriteFileAsync(string path, byte[] content, CancellationToken ct = default)
        {
            await _service.WriteFileToContainerAsync(ID, path, content, ct);
        }

        public async Task<string> ReadFileAsync(string path, CancellationToken ct = default)
        {
            return await _service.ReadFileFromContainerAsync(ID, path, ct);
        }

        public async Task DestroyAsync(CancellationToken ct = default)
        {
            await _service.DestroySandboxAsync(ID, ct);
        }
    }

    public class SandboxService
    {
        private readonly DockerClient _client;
        private readonly string _defaultImage;
        private readonly ILogger<SandboxService> _logger;

        public SandboxService(ILogger<SandboxService> logger, string defaultImage = "python:3.10-slim")
        {
            _logger = logger;
            _defaultImage = string.IsNullOrEmpty(defaultImage) ? "python:3.10-slim" : defaultImage;
            _client = new DockerClientConfiguration(new Uri("npipe://./pipe/docker_engine")).CreateClient();
        }

        public async Task<Sandbox> CreateSandboxAsync(string? image = null, CancellationToken ct = default)
        {
            var targetImage = image ?? _defaultImage;

            try
            {
                await _client.Images.CreateImageAsync(new ImagesCreateParameters
                {
                    FromImage = targetImage
                }, null, new Progress<JSONMessage>(), ct);
            }
            catch (Exception ex)
            {
                _logger.LogWarning(ex, "Failed to pull image {Image}. Trying to use local version.", targetImage);
            }

            var resp = await _client.Containers.CreateContainerAsync(new CreateContainerParameters
            {
                Image = targetImage,
                Cmd = new List<string> { "tail", "-f", "/dev/null" },
                Tty = false,
                AttachStdout = true,
                AttachStderr = true,
                WorkingDir = "/workspace",
                HostConfig = new HostConfig
                {
                    AutoRemove = true,
                    Memory = 512 * 1024 * 1024,
                    NanoCPUs = 500000000
                }
            }, ct);

            await _client.Containers.StartContainerAsync(resp.ID, new ContainerStartParameters(), ct);

            var sandbox = new Sandbox(resp.ID, this);
            await sandbox.ExecAsync("mkdir -p /workspace", ct);

            return sandbox;
        }

        public async Task<(string stdout, string stderr)> ExecInContainerAsync(string id, string command, CancellationToken ct = default)
        {
            var execResp = await _client.Exec.ExecCreateContainerAsync(id, new ContainerExecCreateParameters
            {
                Cmd = new List<string> { "/bin/sh", "-c", command },
                AttachStdout = true,
                AttachStderr = true
            }, ct);

            using var stream = await _client.Exec.StartAndAttachContainerExecAsync(execResp.ID, false, ct);
            var (stdout, stderr) = await stream.ReadOutputToEndAsync(ct);
            return (stdout, stderr);
        }

        public async Task WriteFileToContainerAsync(string id, string path, byte[] content, CancellationToken ct = default)
        {
            // Simple implementation: use base64 and echo to avoid complex tar stream for now
            var base64 = Convert.ToBase64String(content);
            var cmd = $"echo '{base64}' | base64 -d > '{path}'";
            await ExecInContainerAsync(id, cmd, ct);
        }

        public async Task<string> ReadFileFromContainerAsync(string id, string path, CancellationToken ct = default)
        {
            var (stdout, stderr) = await ExecInContainerAsync(id, $"cat '{path}'", ct);
            if (!string.IsNullOrEmpty(stderr))
            {
                throw new Exception($"Failed to read file: {stderr}");
            }
            return stdout;
        }

        public async Task DestroySandboxAsync(string id, CancellationToken ct = default)
        {
            await _client.Containers.StopContainerAsync(id, new ContainerStopParameters { WaitBeforeKillSeconds = 1 }, ct);
        }
    }
}

