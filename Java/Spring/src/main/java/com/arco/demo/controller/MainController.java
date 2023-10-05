package com.arco.demo.controller;

import org.springframework.stereotype.Controller;
import org.springframework.ui.Model;
import org.springframework.web.bind.annotation.GetMapping;
import org.springframework.web.bind.annotation.ResponseBody;

import java.util.HashMap;
import java.util.Map;

@Controller
public class MainController {
    @GetMapping("/version")
    public String version(Model model) {
        model.addAttribute("name", "KimJeongTae");
        model.addAttribute("version", "0.0.1");
        return "version";
    }

    @GetMapping("/healthz")
    @ResponseBody
    public Map<String, String> HealthCheck() {
        Map<String, String> status = new HashMap<>();
        status.put("status", "ok");
        return status;
    }
}
