/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

import { useState } from 'react';

const TABLEAU_10 = [
    '#1f77b4',
    '#ff7f0e',
    '#2ca02c',
    '#d62728',
    '#9467bd',
    '#8c564b',
    '#e377c2',
    '#7f7f7f',
    '#bcbd22',
    '#17becf',
];

const useColor = () => {
    const [colorMap, setColorMap] = useState<{ [key: string]: string }>({});

    const getColor = (key: string) => {
        if (!colorMap[key]) {
            setColorMap((prevColorMap) => ({
                ...prevColorMap,
                [key]: TABLEAU_10[Object.keys(prevColorMap).length % TABLEAU_10.length],
            }));
        }

        return colorMap[key];
    };

    return getColor;
};

export default useColor;
