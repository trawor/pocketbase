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
        With <strong>AgentID</strong>, user will simplely onetap to login WeCom (see
        <a
            href="https://developer.work.weixin.qq.com/document/path/98152"
            target="_blank"
            rel="noopener noreferrer"
        >
            WeCom OAuth docs
        </a>
        ).
    </div>
</div>
