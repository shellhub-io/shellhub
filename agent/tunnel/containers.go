package tunnel

import (
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/labstack/echo/v4"
	"github.com/shellhub-io/shellhub/agent/containers"
	log "github.com/sirupsen/logrus"
)

// streamDelay is the delay between chunks when streaming data to the client.
//
// The idea is to prevent overwhelming the client with too much data at once,
// especially for long-running operations or large data transfers.
const streamDelay = 100 * time.Millisecond

// stream is a helper function to stream data from a reader to the echo context's response.
func stream(reader io.Reader, c echo.Context) error {
	c.Response().Header().Set("Content-Type", "text/event-stream")
	c.Response().Header().Set("Transfer-Encoding", "chunked")
	c.Response().WriteHeader(http.StatusOK)
	c.Response().Flush()

	buffer := make([]byte, 64)
	for {
		n, err := reader.Read(buffer)
		if err != nil {
			log.WithError(err).Error("error reading from stream")

			if err == io.EOF {
				log.Debug("streaming completed")

				return nil
			}

			return err
		}

		if n > 0 {
			if _, err := c.Response().Write(buffer[:n]); err != nil {
				log.WithError(err).Debug("error writing to response")

				return nil
			}

			time.Sleep(streamDelay)
			c.Response().Flush()
		}
	}
}

// ContainersHandler sets up the handlers for managing Docker containers.
func ContainersHandler(e *echo.Echo) {
	log.Info("initializing containers handler")

	c, err := containers.NewDockerConnector()
	if err != nil {
		log.WithError(err).Fatal("failed to create docker connector")
	}

	log.Info("docker connector created successfully")

	logger := func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			start := time.Now()

			err := next(c)

			duration := time.Since(start)
			logger := log.WithFields(log.Fields{
				"method":   c.Request().Method,
				"path":     c.Request().URL.Path,
				"duration": duration,
				"status":   c.Response().Status,
			})

			if err != nil {
				logger.WithError(err).Error("container operation completed with error")
			} else {
				logger.Info("container operation completed successfully")
			}

			return err
		}
	}

	group := e.Group("/containers", logger)

	group.GET("/", func(e echo.Context) error {
		ctx := e.Request().Context()

		logger := log.NewEntry(log.StandardLogger())
		logger.Info("received request to list containers")

		opts := containers.ListOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind list options")

			return err
		}

		logger.WithFields(log.Fields{
			"all":    opts.All,
			"size":   opts.Size,
			"limit":  opts.Limit,
			"latest": opts.Latest,
			"since":  opts.Since,
			"before": opts.Before,
		}).Debug("listing containers with options")

		containers, err := c.List(ctx, opts)
		if err != nil {
			logger.WithError(err).Error("failed to list containers")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContaiersList, err))
		}

		logger.WithField("count", len(containers)).Info("containers listed successfully")

		return e.JSON(http.StatusOK, containers)
	})
	group.GET("/:id", func(e echo.Context) error {
		ctx := e.Request().Context()

		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		logger.Info("received request to get container info")

		if id == "" {
			logger.Error("id received is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, err))
		}

		logger.Debug("fetching container info")

		info, err := c.Info(ctx, id)
		if err != nil {
			logger.WithError(err).Error("failed to get container's info")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContaierGet, err))
		}

		logger.WithFields(log.Fields{
			"status":  info.State.Status,
			"running": info.State.Running,
			"image":   info.Config.Image,
		}).Info("container info retrieved successfully")

		return e.JSON(http.StatusOK, info)
	})
	group.DELETE("/:id", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		logger.Info("received request to delete container")

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		opts := containers.RemoveOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind remove options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidJSON, err))
		}

		logger.WithFields(log.Fields{
			"force":          opts.Force,
			"remove_volumes": opts.RemoveVolumes,
		}).Info("attempting to remove container")

		if err := c.Remove(ctx, id, opts); err != nil {
			logger.WithError(err).Error("failed to remove container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerRemove, err))
		}

		logger.WithFields(log.Fields{
			"force":          opts.Force,
			"remove_volumes": opts.RemoveVolumes,
		}).Info("container removed successfully")

		return e.NoContent(http.StatusNoContent)
	})
	group.POST("/:id/start", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		logger.Info("received request to start container")

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		opts := containers.StartOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind start options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidParam, err))
		}

		logger.WithFields(log.Fields{
			"detach_keys": opts.DetachKeys,
		}).Debug("starting container with options")

		if err := c.Start(ctx, id, opts); err != nil {
			logger.WithError(err).Error("failed to start container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerStart, err))
		}

		logger.Info("container started successfully")

		return e.NoContent(http.StatusNoContent)
	})

	group.POST("/:id/stop", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		logger.Info("received request to stop container")

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		opts := containers.StopOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind stop options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidParam, err))
		}

		logger.WithFields(log.Fields{
			"timeout": opts.Timeout,
		}).Debug("stopping container with options")

		if err := c.Stop(ctx, id, opts); err != nil {
			logger.WithError(err).Error("failed to stop container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerStop, err))
		}

		logger.Info("container stopped successfully")

		return e.NoContent(http.StatusNoContent)
	})

	group.POST("/:id/restart", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		opts := containers.RestartOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind restart options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidParam, err))
		}

		if err := c.Restart(ctx, id, opts); err != nil {
			logger.WithError(err).Error("failed to restart container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerRestart, err))
		}

		logger.Info("container restarted successfully")

		return e.NoContent(http.StatusNoContent)
	})

	group.POST("/:id/kill", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		opts := containers.KillOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind kill options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidJSON, err))
		}

		if err := c.Kill(ctx, id, opts); err != nil {
			logger.WithError(err).Error("failed to kill container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerKill, err))
		}

		logger.WithField("signal", opts.Signal).Info("container killed successfully")

		return e.NoContent(http.StatusNoContent)
	})

	group.POST("/:id/pause", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		if err := c.Pause(ctx, id); err != nil {
			logger.WithError(err).Error("failed to pause container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerPause, err))
		}

		logger.Info("container paused successfully")

		return e.NoContent(http.StatusNoContent)
	})

	group.POST("/:id/unpause", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		if err := c.Unpause(ctx, id); err != nil {
			logger.WithError(err).Error("failed to unpause container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerUnpause, err))
		}

		logger.Info("container unpaused successfully")

		return e.NoContent(http.StatusNoContent)
	})

	// Container inspection and monitoring handlers
	group.GET("/:id/logs", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		opts := containers.LogsOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind logs options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidJSON, err))
		}

		logs, err := c.Logs(ctx, id, opts)
		if err != nil {
			logger.WithError(err).Error("failed to get container logs")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerLogs, err))
		}

		defer logs.Close()

		logger.WithFields(log.Fields{
			"follow":     opts.Follow,
			"stderr":     opts.Stderr,
			"since":      opts.Since,
			"until":      opts.Until,
			"timestamps": opts.Timestamps,
			"tail":       opts.Tail,
		}).Info("container logs requested")

		return stream(logs, e)
	})

	group.GET("/:id/stats", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		logger.Info("received request to get container stats")

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		opts := containers.StatsOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind stats options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidJSON, err))
		}

		logger.WithFields(log.Fields{
			"stream":   opts.Stream,
			"one_shot": opts.OneShot,
		}).Debug("fetching container stats with options")

		stats, err := c.Stats(ctx, id, opts)
		if err != nil {
			logger.WithError(err).Error("failed to get container stats")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerStats, err))
		}
		defer stats.Body.Close()

		logger.WithField("stream", opts.Stream).Info("container stats requested")

		return e.Stream(http.StatusOK, "application/json", stats.Body)
	})

	group.GET("/:id/top", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		opts := containers.TopOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind top options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidJSON, err))
		}

		top, err := c.Top(ctx, id, opts)
		if err != nil {
			logger.WithError(err).Error("failed to get container top")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerTop, err))
		}

		logger.WithField("ps_args", opts.PsArgs).Info("container top requested")

		return e.JSON(http.StatusOK, top)
	})

	group.GET("/:id/changes", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		changes, err := c.Changes(ctx, id)
		if err != nil {
			logger.WithError(err).Error("failed to get container changes")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerChanges, err))
		}

		logger.Info("container changes requested")

		return e.JSON(http.StatusOK, changes)
	})

	group.GET("/:id/export", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		export, err := c.Export(ctx, id)
		if err != nil {
			logger.WithError(err).Error("failed to export container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerExport, err))
		}
		defer export.Close()

		logger.Info("container export requested")
		e.Response().Header().Set("Content-Type", "application/x-tar")

		return e.Stream(http.StatusOK, "application/x-tar", export)
	})

	group.POST("/:id/wait", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		opts := containers.WaitOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind wait options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidJSON, err))
		}

		waitCh, errCh := c.Wait(ctx, id, opts)

		select {
		case result := <-waitCh:
			logger.WithField("condition", opts.Condition).Info("container wait completed")

			return e.JSON(http.StatusOK, result)
		case err := <-errCh:
			logger.WithError(err).Error("failed to wait for container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerWait, err))
		}
	})

	// Container file operations
	group.PUT("/:id/archive", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")
		path := e.QueryParam("path")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id":   id,
			"path": path,
		})

		logger.Info("received request to copy files to container")

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		if path == "" {
			logger.Error("path parameter is required")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidParam, errors.New("path parameter is required")))
		}

		opts := containers.CopyOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind copy options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidJSON, err))
		}

		logger.WithFields(log.Fields{
			"no_overwrite":    opts.NoOverwrite,
			"copy_uid_gid":    opts.CopyUIDGID,
			"allow_overwrite": opts.AllowOverwrite,
		}).Debug("copying files to container with options")

		if err := c.CopyToContainer(ctx, id, path, e.Request().Body, opts); err != nil {
			logger.WithError(err).Error("failed to copy to container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerCopyTo, err))
		}

		logger.Info("container copy to completed successfully")

		return e.NoContent(http.StatusNoContent)
	})

	group.GET("/:id/archive", func(e echo.Context) error {
		ctx := e.Request().Context()

		id := e.Param("id")
		path := e.QueryParam("path")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id":   id,
			"path": path,
		})

		logger.Info("received request to copy files from container")

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		if path == "" {
			logger.Error("path parameter is required")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidParam, errors.New("path parameter is required")))
		}

		logger.Debug("fetching files from container")

		reader, stat, err := c.CopyFromContainer(ctx, id, path)
		if err != nil {
			logger.WithError(err).Error("failed to copy from container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerCopyFrom, err))
		}
		defer reader.Close()

		// Set headers with file stat information
		e.Response().Header().Set("Content-Type", "application/x-tar")
		e.Response().Header().Set("X-Docker-Container-Path-Stat", stat.Name)

		logger.WithFields(log.Fields{
			"stat_name": stat.Name,
			"stat_size": stat.Size,
			"stat_mode": stat.Mode,
		}).Info("container copy from completed successfully")

		return e.Stream(http.StatusOK, "application/x-tar", reader)
	})

	// Container management operations
	group.POST("/:id/rename", func(e echo.Context) error {
		ctx := e.Request().Context()

		id := e.Param("id")
		name := e.QueryParam("name")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id":   id,
			"name": name,
		})

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		if name == "" {
			logger.Error("name parameter is required")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidParam, errors.New("name parameter is required")))
		}

		opts := containers.RenameOptions{Name: name}
		if err := c.Rename(ctx, id, opts); err != nil {
			logger.WithError(err).Error("failed to rename container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerRename, err))
		}

		logger.Info("container renamed successfully")

		return e.NoContent(http.StatusNoContent)
	})

	group.POST("/:id/update", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		opts := containers.UpdateOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind update options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidJSON, err))
		}

		result, err := c.Update(ctx, id, opts)
		if err != nil {
			logger.WithError(err).Error("failed to update container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerUpdate, err))
		}

		logger.Info("container updated successfully")

		return e.JSON(http.StatusOK, result)
	})

	group.POST("/:id/resize", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		if id == "" {
			logger.Error("container id is invalid")
			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		opts := containers.ResizeOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind resize options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidJSON, err))
		}

		if err := c.Resize(ctx, id, opts); err != nil {
			logger.WithError(err).Error("failed to resize container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerResize, err))
		}

		logger.WithFields(log.Fields{
			"height": opts.Height,
			"width":  opts.Width,
		}).Info("container resized successfully")

		return e.NoContent(http.StatusNoContent)
	})

	group.POST("/:id/commit", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		opts := containers.CommitOptions{}
		if err := e.Bind(&opts); err != nil {
			logger.WithError(err).Error("failed to bind commit options")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainerInvalidJSON, err))
		}

		result, err := c.Commit(ctx, id, opts)
		if err != nil {
			logger.WithError(err).Error("failed to commit container")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerCommit, err))
		}

		logger.WithFields(log.Fields{
			"repo":    opts.Repo,
			"tag":     opts.Tag,
			"comment": opts.Comment,
			"author":  opts.Author,
			"pause":   opts.Pause,
		}).Info("container committed successfully")

		return e.JSON(http.StatusOK, result)
	})

	// Health check endpoint
	group.GET("/:id/health", func(e echo.Context) error {
		ctx := e.Request().Context()
		id := e.Param("id")

		logger := log.NewEntry(log.StandardLogger()).WithFields(log.Fields{
			"id": id,
		})

		logger.Info("received request to check container health")

		if id == "" {
			logger.Error("container id is invalid")

			return e.JSON(http.StatusBadRequest, errors.Join(ErrAgentContainersIDInvalid, errors.New("container id is required")))
		}

		logger.Debug("fetching container info for health check")

		info, err := c.Info(ctx, id)
		if err != nil {
			logger.WithError(err).Error("failed to get container info for health check")

			return e.JSON(http.StatusInternalServerError, errors.Join(ErrAgentContainerHealth, err))
		}

		healthStatus := map[string]any{
			"Status":     info.State.Status,
			"Running":    info.State.Running,
			"Paused":     info.State.Paused,
			"Restarting": info.State.Restarting,
			"Dead":       info.State.Dead,
			"StartedAt":  info.State.StartedAt,
			"FinishedAt": info.State.FinishedAt,
		}

		if info.State.Health != nil {
			healthStatus["Health"] = info.State.Health

			logger.WithFields(log.Fields{
				"health_status":  info.State.Health.Status,
				"failing_streak": info.State.Health.FailingStreak,
			}).Debug("container health check details")
		}

		logger.WithFields(log.Fields{
			"status":     info.State.Status,
			"running":    info.State.Running,
			"paused":     info.State.Paused,
			"restarting": info.State.Restarting,
			"dead":       info.State.Dead,
		}).Info("container health check completed successfully")

		return e.JSON(http.StatusOK, healthStatus)
	})
}
