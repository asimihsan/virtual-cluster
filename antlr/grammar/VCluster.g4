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

vclusterConfig: configEntry+ EOF;

configEntry: serviceEntry
           | dependencyEntry
           ;

serviceEntry: 'service' serviceName '{' serviceConfigItem+ '}';

dependencyEntry: 'dependency' dependencyName '{' dependencyConfigItem+ '}';

serviceName: IDENTIFIER;

dependencyName: IDENTIFIER;

serviceConfigItem: 'repository' ':' STRING_LITERAL ';'?  # serviceConfigRepository
                 | 'branch' ':' STRING_LITERAL ';'?      # serviceConfigBranch
                 | 'tag' ':' STRING_LITERAL ';'?         # serviceConfigTag
                 | 'commit' ':' STRING_LITERAL ';'?      # serviceConfigCommit
                 | 'directory' ':' STRING_LITERAL ';'?   # serviceConfigDirectory
                 | 'health_check' '{' healthCheck+ '}'   # serviceConfigHealthCheck
                 | 'dependency' ':' IDENTIFIER ';'?      # serviceConfigDependency
                 | 'run_commands' ':' '[' STRING_LITERAL (',' STRING_LITERAL)* ','? ']' ';'?  # serviceConfigRunCommands
                 ;

dependencyConfigItem: 'health_check' '{' healthCheck+ '}'   # dependencyConfigHealthCheck
                     | 'dependency' ':' IDENTIFIER ';'?     # dependencyConfigDependency
                     ;

healthCheck: 'endpoint' ':' STRING_LITERAL ';'?         # healthCheckEndpoint
           ;

IDENTIFIER: [a-zA-Z_][a-zA-Z_0-9]*;

STRING_LITERAL: '"' (ESC|.)*? '"' | [a-zA-Z_][a-zA-Z_0-9.-]*;
fragment
ESC : '\\"' | '\\\\' ; // 2-char sequences \" and \\

WS: [ \t\r\n]+ -> skip;
C_BLOCK_COMMENT: '/*' .*? '*/' -> skip;
CPP_LINE_COMMENT: '//' ~[\r\n]* -> skip;
PYTHON_LINE_COMMENT: '#' ~[\r\n]* -> skip;
