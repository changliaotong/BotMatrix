using System;
using System.ComponentModel.DataAnnotations.Schema;

namespace BotWorker.Modules.AI.Models
{
    [Table("ai_llm_call_logs")]
    public class LLMCallLog
    {
        public long Id { get; set; }
        
        [System.ComponentModel.DataAnnotations.Schema.Column("task_step_id")]
        public long? TaskStepId { get; set; }
        
        [System.ComponentModel.DataAnnotations.Schema.Column("agent_id")]
        public long? AgentId { get; set; }
        
        [System.ComponentModel.DataAnnotations.Schema.Column("model_id")]
        public long? ModelId { get; set; }
        
        [System.ComponentModel.DataAnnotations.Schema.Column("prompt_tokens")]
        public int PromptTokens { get; set; }
        
        [System.ComponentModel.DataAnnotations.Schema.Column("completion_tokens")]
        public int CompletionTokens { get; set; }
        
        [System.ComponentModel.DataAnnotations.Schema.Column("total_cost")]
        public decimal TotalCost { get; set; }
        
        [System.ComponentModel.DataAnnotations.Schema.Column("latency_ms")]
        public int LatencyMs { get; set; }
        
        [System.ComponentModel.DataAnnotations.Schema.Column("is_success")]
        public bool IsSuccess { get; set; } = true;
        
        [System.ComponentModel.DataAnnotations.Schema.Column("raw_request")]
        public string? RawRequest { get; set; } // JSONB
        
        [System.ComponentModel.DataAnnotations.Schema.Column("raw_response")]
        public string? RawResponse { get; set; } // JSONB
        
        [System.ComponentModel.DataAnnotations.Schema.Column("created_at")]
        public DateTime CreatedAt { get; set; } = DateTime.UtcNow;

        [System.ComponentModel.DataAnnotations.Schema.Column("updated_at")]
        public DateTime UpdatedAt { get; set; } = DateTime.UtcNow;
    }
}
