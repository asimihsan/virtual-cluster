/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

grammar Services;

servicesConfig: serviceEntry+ EOF;

serviceEntry: 'service' serviceName '{' serviceConfigItem+ '}';

serviceName: IDENTIFIER;

serviceConfigItem: 'repository' ':' STRING_LITERAL ';'
                 | 'branch' ':' IDENTIFIER ';'
                 | 'tag' ':' IDENTIFIER ';'
                 | 'commit' ':' IDENTIFIER ';'
                 | 'directory' ':' STRING_LITERAL ';';

IDENTIFIER: [a-zA-Z_][a-zA-Z_0-9]*;
STRING_LITERAL: '"' ~["]* '"';
WS: [ \t\r\n]+ -> skip;
C_BLOCK_COMMENT: '/*' .*? '*/' -> skip;
CPP_LINE_COMMENT: '//' ~[\r\n]* -> skip;
PYTHON_LINE_COMMENT: '#' ~[\r\n]* -> skip;
