/*
 * Copyright (c) 2023 Asim Ihsan.
 *
 * This Source Code Form is subject to the terms of the Mozilla Public
 * License, v. 2.0. If a copy of the MPL was not distributed with this
 * file, You can obtain one at http://mozilla.org/MPL/2.0/.
 *
 * SPDX-License-Identifier: MPL-2.0
 */

grammar VCluster;

vclusterConfig: serviceConfig+ EOF;

serviceConfig: 'service' serviceName '{' configItem+ '}';

serviceName: IDENTIFIER;

configItem: 'dependency' '{' dependency+ '}'            # dependencyConfigItem
          | 'health_check' '{' healthCheck+ '}'         # healthCheckConfigItem
          | 'startup_sequence' '{' startupSequence+ '}' # startupSequenceConfigItem
          ;

dependency: 'name' ':' IDENTIFIER ';'?                  # dependencyName
          ;

healthCheck: 'endpoint' ':' STRING_LITERAL ';'?         # endpointHealthCheck
           ;

startupSequence: 'command' ':' STRING_LITERAL ';'?      # commandStartupSequence
               ;

IDENTIFIER: [a-zA-Z_][a-zA-Z_0-9]*;

STRING_LITERAL: '"' (ESC|.)*? '"' | [a-zA-Z_][a-zA-Z_0-9.-]*;
fragment
ESC : '\\"' | '\\\\' ; // 2-char sequences \" and \\

INTEGER_LITERAL: [0-9]+;
WS: [ \t\r\n]+ -> skip;
C_BLOCK_COMMENT: '/*' .*? '*/' -> skip;
CPP_LINE_COMMENT: '//' ~[\r\n]* -> skip;
PYTHON_LINE_COMMENT: '#' ~[\r\n]* -> skip;
