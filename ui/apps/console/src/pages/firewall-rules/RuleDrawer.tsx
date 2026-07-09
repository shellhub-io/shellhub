import { useWatch } from "react-hook-form";
import {
  useCreateFirewallRule,
  useUpdateFirewallRule,
} from "@/hooks/useFirewallRuleMutations";
import type { FirewallRulesResponse } from "@/client";
import RadioCard from "@/components/common/fields/RadioCard";
import {
  UserGroupIcon,
  UserIcon as UserIconHero,
  ClipboardDocumentListIcon,
  TagIcon as TagIconHero,
  GlobeAltIcon,
  ComputerDesktopIcon,
  CheckCircleIcon,
  NoSymbolIcon,
} from "@heroicons/react/24/outline";
import { DevicesIcon as DevicesIconComponent } from "@shellhub/design-system/primitives";
import FormDrawer from "@/components/common/FormDrawer";
import {
  FormInputField,
  FormNumericInput,
  FormRadioGroupField,
  FormTagsSelector,
  FormToggleField,
} from "@/components/common/fields/rhf";
import { useDrawerForm } from "@/hooks/useDrawerForm";
import {
  ruleSchema,
  DEFAULT_VALUES,
  buildRuleBody,
  buildRuleDefaults,
  type RuleFormValues,
} from "./ruleSchema";

const UsersIcon = <UserGroupIcon className="w-4 h-4" />;
const UserIcon = <UserIconHero className="w-4 h-4" />;
const DevicesIcon = <DevicesIconComponent className="w-4 h-4" />;
const HostnameIcon = <ClipboardDocumentListIcon className="w-4 h-4" />;
const TagIcon = <TagIconHero className="w-4 h-4" />;
const GlobeIcon = <GlobeAltIcon className="w-4 h-4" />;
const IpIcon = <ComputerDesktopIcon className="w-4 h-4" />;

export default function RuleDrawer({
  open,
  editRule,
  onClose,
}: {
  open: boolean;
  editRule: FirewallRulesResponse | null;
  onClose: () => void;
}) {
  const createRule = useCreateFirewallRule();
  const updateRule = useUpdateFirewallRule();
  const isEdit = !!editRule;

  const form = useDrawerForm(
    open,
    ruleSchema,
    editRule ? buildRuleDefaults(editRule) : DEFAULT_VALUES,
  );
  const { control, setError } = form;

  const sourceIpOption = useWatch({ control, name: "sourceIpOption" });
  const usernameOption = useWatch({ control, name: "usernameOption" });
  const filterOption = useWatch({ control, name: "filterOption" });

  const onSubmit = async (values: RuleFormValues) => {
    const body = buildRuleBody(values);
    try {
      if (isEdit && editRule) {
        await updateRule.mutateAsync({ path: { id: editRule.id }, body });
      } else {
        await createRule.mutateAsync({ body });
      }
      onClose();
    } catch (err: unknown) {
      setError("root", {
        message:
          err instanceof Error
            ? err.message
            : `Failed to ${isEdit ? "update" : "create"} rule`,
      });
    }
  };

  return (
    <FormDrawer
      form={form}
      onSubmit={onSubmit}
      open={open}
      onClose={onClose}
      title={isEdit ? "Edit Firewall Rule" : "New Firewall Rule"}
      submitLabel={isEdit ? "Save Changes" : "Create Rule"}
    >
      <FormToggleField name="active" control={control} label="Status" />

      <FormNumericInput
        name="priority"
        control={control}
        id="rule-priority"
        label="Priority"
        placeholder="e.g. 100"
        hint="Higher values are evaluated first. Must be greater than 0."
      />

      <FormRadioGroupField<RuleFormValues, "allow" | "deny">
        name="action"
        control={control}
        label="Policy"
      >
        <RadioCard
          value="allow"
          icon={<CheckCircleIcon className="w-4 h-4" />}
          label="Allow"
          description="Permit the connection when this rule matches."
        />
        <RadioCard
          value="deny"
          icon={<NoSymbolIcon className="w-4 h-4" />}
          label="Deny"
          description="Block the connection when this rule matches."
        />
      </FormRadioGroupField>

      <div>
        <FormRadioGroupField<RuleFormValues, "all" | "restrict">
          name="sourceIpOption"
          control={control}
          label="Source IP"
        >
          <RadioCard
            value="all"
            icon={GlobeIcon}
            label="All source IPs"
            description="Match connections from any IP address."
          />
          <RadioCard
            value="restrict"
            icon={IpIcon}
            label="Restrict with regexp"
            description="Match connections from IPs matching a pattern."
          />
        </FormRadioGroupField>
        {sourceIpOption === "restrict" && (
          <div className="mt-2">
            <FormInputField
              name="sourceIp"
              control={control}
              id="rule-source-ip-pattern"
              label="Source IP pattern"
              hideLabel
              placeholder="e.g. 192\.168\.1\..*"
              variant="mono"
            />
          </div>
        )}
      </div>

      <div>
        <FormRadioGroupField<RuleFormValues, "all" | "restrict">
          name="usernameOption"
          control={control}
          label="Username"
        >
          <RadioCard
            value="all"
            icon={UsersIcon}
            label="All users"
            description="Match connections for any username."
          />
          <RadioCard
            value="restrict"
            icon={UserIcon}
            label="Restrict with regexp"
            description="Match connections for usernames matching a pattern."
          />
        </FormRadioGroupField>
        {usernameOption === "restrict" && (
          <div className="mt-2">
            <FormInputField
              name="username"
              control={control}
              id="rule-username-pattern"
              label="Username pattern"
              hideLabel
              placeholder="e.g. root"
              variant="mono"
            />
          </div>
        )}
      </div>

      <div>
        <FormRadioGroupField<RuleFormValues, "all" | "hostname" | "tags">
          name="filterOption"
          control={control}
          label="Device filter"
        >
          <RadioCard
            value="all"
            icon={DevicesIcon}
            label="All devices"
            description="Match connections to any device in the namespace."
          />
          <RadioCard
            value="hostname"
            icon={HostnameIcon}
            label="Filter by hostname"
            description="Restrict to devices matching a hostname pattern."
          />
          <RadioCard
            value="tags"
            icon={TagIcon}
            label="Filter by tags"
            description="Restrict to devices matching specific tags."
          />
        </FormRadioGroupField>
        {filterOption === "hostname" && (
          <div className="mt-2">
            <FormInputField
              name="hostname"
              control={control}
              id="rule-hostname-pattern"
              label="Hostname pattern"
              hideLabel
              placeholder="e.g. web-.*"
              variant="mono"
            />
          </div>
        )}
        {filterOption === "tags" && (
          <div className="mt-2">
            <FormTagsSelector
              name="tags"
              control={control}
              id="rule-filter-tags"
              label="Filter by tags"
            />
          </div>
        )}
      </div>
    </FormDrawer>
  );
}
