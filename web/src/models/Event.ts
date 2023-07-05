export interface LogEvent {
    id: number;
    type: string;
    timestamp: string;
    process_name: string;
    output_type: string;
    content: string;
}

export interface HttpRequestEvent {
    id: number;
    type: string;
    timestamp: string;
    process_name: string;
    method: string;
    url: string;
    headers: string;
    body: string;
}

export interface KafkaMessageEvent {
    id: number;
    type: string;
    timestamp: string;
    process_name: string;
    broker_name: string;
    topic_name: string;
    message_key: string;
    message_value: string;
}
