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

serviceConfigItem: 'repository' keyValueDelimiter STRING_LITERAL ';'?  # serviceConfigRepository
                 | 'branch' keyValueDelimiter STRING_LITERAL ';'?      # serviceConfigBranch
                 | 'tag' keyValueDelimiter STRING_LITERAL ';'?         # serviceConfigTag
                 | 'commit' keyValueDelimiter STRING_LITERAL ';'?      # serviceConfigCommit
                 | 'directory' keyValueDelimiter STRING_LITERAL ';'?   # serviceConfigDirectory
                 | 'health_check' '{' healthCheck+ '}'   # serviceConfigHealthCheck
                 | 'dependency' keyValueDelimiter IDENTIFIER ';'?      # serviceConfigDependency
                 | 'service_port' keyValueDelimiter PORT ';'?          # serviceConfigPort
                 | 'proxy_port' keyValueDelimiter PORT ';'?            # serviceConfigProxyPort
                 | 'run_commands' keyValueDelimiter '[' STRING_LITERAL (',' STRING_LITERAL)* ','? ']' ';'?  # serviceConfigRunCommands
                 ;

dependencyConfigItem: 'health_check' '{' healthCheck+ '}'   # dependencyConfigHealthCheck
                     | 'dependency' keyValueDelimiter IDENTIFIER ';'?     # dependencyConfigDependency
                     | 'managed_kafka' '{' managedKafkaConfigItem+ '}'    # dependencyConfigManagedKafka
                     ;

healthCheck: 'endpoint' keyValueDelimiter STRING_LITERAL ';'?         # healthCheckEndpoint
           ;

managedKafkaConfigItem: 'port' keyValueDelimiter PORT ';'?            # managedKafkaConfigPort
                       ;

keyValueDelimiter: ':' | '=';

IDENTIFIER: [a-zA-Z_][a-zA-Z_0-9]*;
STRING_LITERAL: '"' (ESC|.)*? '"' | [a-zA-Z_][a-zA-Z_0-9.-]*;
fragment
ESC : '\\"' | '\\\\' ; // 2-char sequences \" and \\
PORT : [0-9]+;

WS: [ \t\r\n]+ -> skip;
C_BLOCK_COMMENT: '/*' .*? '*/' -> skip;
CPP_LINE_COMMENT: '//' ~[\r\n]* -> skip;
PYTHON_LINE_COMMENT: '#' ~[\r\n]* -> skip;
