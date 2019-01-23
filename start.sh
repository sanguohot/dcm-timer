#!/usr/bin/env bash
export DCM_TIMER_PATH=/opt/dcm-timer
export DCM_TIMER_TYPE=production
#nohup /opt/dcm-timer/dcm-timer > /opt/dcm-timer/nohup.out 2>&1 &
/opt/dcm-timer/dcm-timer -d true
