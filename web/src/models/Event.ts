/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

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

export interface HttpResponseEvent {
    id: number;
    type: string;
    timestamp: string;
    process_name: string;
    status_code: number;
    headers: string;
    body: string;
    http_request: HttpRequestEvent;
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
