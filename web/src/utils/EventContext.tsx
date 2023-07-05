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
import { LogEvent, HttpRequestEvent, KafkaMessageEvent } from '../models/Event';

type Event = LogEvent | HttpRequestEvent | KafkaMessageEvent;

interface EventContextValue {
    events: Event[];
    addEvent: (event: Event) => void;
}

const EventContext = React.createContext<EventContextValue | undefined>(undefined);

// @ts-ignore
export const EventProvider: React.FC = ({ children }) => {
    const [events, setEvents] = React.useState<Event[]>([]);
    const eventMap = React.useRef(new Map<string, Event>());

    const addEvent = (event: Event) => {
        const eventKey = `${event.type}-${event.id}`;

        if (!eventMap.current.has(eventKey)) {
            eventMap.current.set(eventKey, event);
            setEvents((prevEvents) => [...prevEvents, event].sort((a, b) => a.timestamp.localeCompare(b.timestamp)));
        }
    };

    return (
        <EventContext.Provider value={{ events, addEvent }}>
            {children}
        </EventContext.Provider>
    );
};

export const useEventContext = () => {
    const context = React.useContext(EventContext);
    if (!context) {
        throw new Error('useEventContext must be used within an EventProvider');
    }
    return context;
};
