import { useState, useEffect, FormEvent } from "react";
import {
  UserGroupIcon,
  UserIcon,
  TagIcon,
  ExclamationCircleIcon,
  ClipboardDocumentListIcon,
} from "@heroicons/react/24/outline";
import { DevicesIcon } from "../../components/icons";
import { usePublicKeysStore } from "../../stores/publicKeysStore";
import { PublicKey, PublicKeyFilter } from "../../types/publicKey";
import { isPublicKeyValid } from "../../utils/sshKeys";
import RadioCard from "../../components/common/RadioCard";
import TagsSelector from "../../components/common/TagsSelector";
import Drawer from "../../components/common/Drawer";
import axios from "axios";
import { LABEL, INPUT, INPUT_MONO } from "../../utils/styles";
import KeyDataInput from "./KeyDataInput";

/* --- Drawer --- */
function KeyDrawer({
  open,
  editKey,
  onClose,
}: {
  open: boolean;
  editKey: PublicKey | null;
  onClose: () => void;
}) {
  const { create, update } = usePublicKeysStore();
  const isEdit = !!editKey;

  const decodedKeyData = editKey
    ? (() => {
        try {
          return atob(editKey.data);
        } catch {
          return editKey.data;
        }
      })()
    : "";

  const [name, setName] = useState("");
  const [keyData, setKeyData] = useState("");
  const [keyError, setKeyError] = useState<string | null>(null);
  const [usernameOption, setUsernameOption] = useState<"all" | "username">(
    "all",
  );
  const [username, setUsername] = useState("");
  const [filterOption, setFilterOption] = useState<"all" | "hostname" | "tags">(
    "all",
  );
  const [hostname, setHostname] = useState("");
  const [selectedTags, setSelectedTags] = useState<string[]>([]);
  const [submitting, setSubmitting] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // Populate on open / editKey change
  useEffect(() => {
    if (!open) return;
    if (editKey) {
      setName(editKey.name);
      setKeyData(decodedKeyData);
      setUsernameOption(editKey.username === ".*" ? "all" : "username");
      setUsername(editKey.username === ".*" ? "" : editKey.username);
      if (editKey.filter.tags && editKey.filter.tags.length > 0) {
        setFilterOption("tags");
        setSelectedTags(editKey.filter.tags);
      } else if (editKey.filter.hostname && editKey.filter.hostname !== ".*") {
        setFilterOption("hostname");
        setHostname(editKey.filter.hostname);
      } else {
        setFilterOption("all");
        setHostname("");
        setSelectedTags([]);
      }
      setKeyError(null);
      setError(null);
    } else {
      setName("");
      setKeyData("");
      setKeyError(null);
      setUsernameOption("all");
      setUsername("");
      setFilterOption("all");
      setHostname("");
      setSelectedTags([]);
      setError(null);
    }
  }, [open, editKey, decodedKeyData]);

  const handleKeyDataChange = (v: string) => {
    setKeyData(v);
    if (v && !isPublicKeyValid(v))
      setKeyError("This is not a valid public key.");
    else setKeyError(null);
  };

  const handleFileName = (filename: string) => {
    if (!name) setName(filename || "Imported Public Key");
  };

  const buildFilter = (): PublicKeyFilter => {
    if (filterOption === "hostname" && hostname) return { hostname };
    if (filterOption === "tags" && selectedTags.length > 0)
      return { tags: selectedTags };
    return { hostname: ".*" };
  };

  const tagError =
    selectedTags.length > 3
      ? "You can select up to 3 tags"
      : filterOption === "tags" && selectedTags.length === 0
        ? "Select at least one tag"
        : undefined;

  const confirmDisabled = isEdit
    ? !name.trim() ||
      (usernameOption === "username" && !username.trim()) ||
      (filterOption === "hostname" && !hostname.trim()) ||
      (filterOption === "tags" &&
        (selectedTags.length === 0 || selectedTags.length > 3))
    : !name.trim() ||
      !keyData.trim() ||
      !!keyError ||
      (usernameOption === "username" && !username.trim()) ||
      (filterOption === "hostname" && !hostname.trim()) ||
      (filterOption === "tags" &&
        (selectedTags.length === 0 || selectedTags.length > 3));

  const handleSubmit = async (e: FormEvent) => {
    e.preventDefault();
    if (confirmDisabled) return;
    setError(null);
    setSubmitting(true);
    try {
      if (isEdit && editKey) {
        await update(editKey.fingerprint, {
          name: name.trim(),
          username: usernameOption === "all" ? ".*" : username.trim(),
          filter: buildFilter(),
        });
      } else {
        await create({
          name: name.trim(),
          data: btoa(keyData.trim()),
          username: usernameOption === "all" ? ".*" : username.trim(),
          filter: buildFilter(),
        });
      }
      onClose();
    } catch (err: unknown) {
      if (!isEdit && axios.isAxiosError(err) && err.response?.status === 409) {
        setKeyError("This public key already exists.");
      } else {
        setError(
          err instanceof Error
            ? err.message
            : `Failed to ${isEdit ? "update" : "create"} public key`,
        );
      }
    } finally {
      setSubmitting(false);
    }
  };

  return (
    <Drawer
      open={open}
      onClose={onClose}
      title={isEdit ? "Edit Public Key" : "New Public Key"}
      footer={
        <>
          <button
            type="button"
            onClick={onClose}
            className="px-4 py-2.5 text-sm font-medium text-text-secondary hover:text-text-primary rounded-lg hover:bg-hover-subtle transition-colors"
          >
            Cancel
          </button>
          <button
            type="submit"
            onClick={handleSubmit}
            disabled={submitting || confirmDisabled}
            className="px-5 py-2.5 bg-primary hover:bg-primary-600 text-white rounded-lg text-sm font-semibold disabled:opacity-dim disabled:cursor-not-allowed transition-all"
          >
            {submitting ? (
              <span className="flex items-center gap-2">
                <span className="w-3.5 h-3.5 border-2 border-white/30 border-t-white rounded-full animate-spin" />
                Saving...
              </span>
            ) : isEdit ? (
              "Save Changes"
            ) : (
              "Create Key"
            )}
          </button>
        </>
      }
    >
      <form onSubmit={handleSubmit} className="space-y-5">
        {/* Name */}
        <div>
          <label className={LABEL}>Name</label>
          <input
            type="text"
            value={name}
            onChange={(e) => setName(e.target.value)}
            placeholder="Name used to identify the public key"
            autoFocus={open}
            className={INPUT}
          />
        </div>

        {/* Username access */}
        <div>
          <label className={LABEL}>Username access</label>
          <div className="space-y-2">
            <RadioCard
              selected={usernameOption === "all"}
              onClick={() => setUsernameOption("all")}
              icon={<UserGroupIcon className="w-4 h-4" />}
              label="Allow any user"
              description="The key will work for all usernames on the device."
            />
            <RadioCard
              selected={usernameOption === "username"}
              onClick={() => setUsernameOption("username")}
              icon={<UserIcon className="w-4 h-4" />}
              label="Restrict by username"
              description="Only allow connections matching a username pattern."
            />
          </div>
          {usernameOption === "username" && (
            <input
              type="text"
              value={username}
              onChange={(e) => setUsername(e.target.value)}
              placeholder="e.g. root"
              className={`${INPUT_MONO} mt-2`}
            />
          )}
        </div>

        {/* Device access */}
        <div>
          <label className={LABEL}>Device access</label>
          <div className="space-y-2">
            <RadioCard
              selected={filterOption === "all"}
              onClick={() => setFilterOption("all")}
              icon={<DevicesIcon className="w-4 h-4" />}
              label="All devices"
              description="The key will be accepted by any device in the namespace."
            />
            <RadioCard
              selected={filterOption === "hostname"}
              onClick={() => setFilterOption("hostname")}
              icon={<ClipboardDocumentListIcon className="w-4 h-4" />}
              label="Filter by hostname"
              description="Restrict access using a regexp pattern for hostname."
            />
            <RadioCard
              selected={filterOption === "tags"}
              onClick={() => setFilterOption("tags")}
              icon={<TagIcon className="w-4 h-4" />}
              label="Filter by tags"
              description="Restrict access to devices matching specific tags."
            />
          </div>
          {filterOption === "hostname" && (
            <input
              type="text"
              value={hostname}
              onChange={(e) => setHostname(e.target.value)}
              placeholder="e.g. .*"
              className={`${INPUT_MONO} mt-2`}
            />
          )}
          {filterOption === "tags" && (
            <div className="mt-2">
              <TagsSelector
                selected={selectedTags}
                onChange={setSelectedTags}
                error={tagError}
              />
            </div>
          )}
        </div>

        {/* Public Key Data */}
        <KeyDataInput
          value={keyData}
          onChange={handleKeyDataChange}
          error={keyError || undefined}
          disabled={isEdit}
          onFileName={handleFileName}
        />

        {/* Error */}
        {error && (
          <p className="text-xs font-mono text-accent-red flex items-center gap-1.5">
            <ExclamationCircleIcon
              className="w-3.5 h-3.5 shrink-0"
              strokeWidth={2}
            />
            {error}
          </p>
        )}
      </form>
    </Drawer>
  );
}

export default KeyDrawer;
