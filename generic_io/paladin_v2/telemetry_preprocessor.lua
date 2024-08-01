function preprocess_telemetry(telemetry)
    ret = {}
    ret["timestamp"] = telemetry["timestamp"]

    brain_stats = telemetry["brain"]
    ret["brain_power"] = brain_stats["power"]

    arm_stats = telemetry["arms"]
    if arm_stats ~= nil then
        arm_state = bit.band(arm_stats, 0xFF)
        if arm_state == 0 then
            ret["arms_state"] = "lowered"
        else
            ret["arms_state"] = "raised"
        end
    end


    if telemetry["errors"] ~= nil then
        ret["alerts"] = {}
        ret["alert_details"] = {}
        for i, v in ipairs(telemetry["errors"]) do
            table.insert(ret["alerts"], v)
            if telemetry["errors_description"] ~= nil then
                ret["alert_details"][v] = telemetry["errors_description"][v]
            end
        end
    end

    if telemetry["alerts_nil"] then
        ret["alerts"] = {}
        ret["alert_details"] = {}
    end

    if telemetry["arms_nil"] then
        ret["arms"] = {}
    end

    return ret
end
