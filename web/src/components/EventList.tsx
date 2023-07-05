/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

import React from 'react';
import { useEventContext } from '../utils/EventContext';
import useEventService from '../services/useEventService';
import useColor from '../utils/useColor';
import {HttpRequestEvent, HttpResponseEvent, KafkaMessageEvent, LogEvent} from "../models/Event";

import './EventList.css';


const EventList: React.FC = () => {
    const { events } = useEventContext();
    const getColor = useColor();

    useEventService();

    const getEventContent = (event: LogEvent | HttpRequestEvent | HttpResponseEvent | KafkaMessageEvent) => {
        switch (event.type) {
            case 'log':
                const logEvent = event as LogEvent;
                return logEvent.content?.substring(0, 100);
            case 'http_request':
                const httpRequestEvent = event as HttpRequestEvent;
                // if body is null or empty string, use <no body> else get it and truncate
                var body = httpRequestEvent.body;
                if (body === null || body === '') {
                    body = '<no body>';
                }
                return `${httpRequestEvent.method} - ${httpRequestEvent.url} - ${body?.substring(0, 100)}`;
            case 'http_response':
                const httpResponseEvent = event as HttpResponseEvent;
                const httpRequestEvent2 = httpResponseEvent.http_request;
                return `${httpRequestEvent2.method} - ${httpRequestEvent2.url} - ${httpResponseEvent.status_code} - ${httpResponseEvent.body?.substring(0, 100)}`;
            case 'kafka_message':
                const kafkaMessageEvent = event as KafkaMessageEvent;
                return `${kafkaMessageEvent.broker_name} - ${kafkaMessageEvent.topic_name} - ${kafkaMessageEvent.message_value?.substring(0, 100)}`;
            default:
                return '';
        }
    };

    const getProcessName = (event: LogEvent | HttpRequestEvent | HttpResponseEvent | KafkaMessageEvent) => {
        switch (event.type) {
            case 'log':
                const logEvent = event as LogEvent;
                return logEvent.process_name;
            case 'http_request':
                const httpRequestEvent = event as HttpRequestEvent;
                return httpRequestEvent.process_name;
            case 'http_response':
                const httpResponseEvent = event as HttpResponseEvent;
                return httpResponseEvent.process_name;
            case 'kafka_message':
                return '';
            default:
                return '';
        }
    }

    return (
        <table className="event-table">
            <tbody>
            {events.map((event, index) => {
                const backgroundColor = getColor(event.process_name || 'kafka_message');
                return (
                    <tr key={index} style={{ backgroundColor }}>
                        <td><div>{event.timestamp}</div></td>
                        <td><div>{getProcessName(event)}</div></td>
                        <td><div>{event.type}</div></td>
                        <td><div>{getEventContent(event)}</div></td>
                    </tr>
                );
            })}
            </tbody>
        </table>

    );
};

export default EventList;
