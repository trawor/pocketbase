<script>
    import Field from "@/components/base/Field.svelte";
    export let key = "";
    export let config = {};
    if (!config.extra) {
        config.extra = {};
    }
</script>

<Field class="form-field" name="{key}.extra.agentID" let:uniqueId>
    <label for={uniqueId}>AgentID</label>
    <input id={uniqueId} bind:value={config.extra.agentID} required="false" />
</Field>

<div class="alert alert-info">
    <div class="icon">
        <i class="ri-information-line" />
    </div>
    <div class="content">
        Note that the Lark user's <strong>Union ID</strong> will be used for the association with the
        PocketBase user (see
        <a
            href="https://open.feishu.cn/document/platform-overveiw/basic-concepts/user-identity-introduction/introduction#3f2d4b63"
            target="_blank"
            rel="noopener noreferrer"
        >
            Different Types of Lark IDs
        </a>
        ).
    </div>
</div>
